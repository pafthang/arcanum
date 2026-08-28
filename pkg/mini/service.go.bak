package mini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
)

// AddService adds a microservice.
// It will enable internal common services (PING, STATS and INFO).
// Request handlers have to be registered separately using Service.AddEndpoint.
// A service name, version and Endpoint configuration are required to add a service.
// AddService returns a [Service] interface, allowing service management.
// Each service is assigned a unique ID.
func AddService(nc *nats.Conn, config Config) (Service, error) {
	if err := config.valid(); err != nil {
		return nil, err
	}

	if config.Metadata == nil {
		config.Metadata = map[string]string{}
	}

	id := nuid.Next()
	baseCtx, baseCancel := context.WithCancel(context.Background())
	svc := &service{
		Config: config,
		nc:     nc,
		id:     id,
		asyncDispatcher: asyncCallbacksHandler{
			cbQueue: make(chan func(), 100),
		},
		verbSubs:   make(map[string]*nats.Subscription),
		endpoints:  make([]*Endpoint, 0),
		baseCtx:    baseCtx,
		baseCancel: baseCancel,
	}

	// Add connection event (closed, error) wrapper handlers. If the service has
	// custom callbacks, the events are queued and invoked by the same
	// goroutine, starting now.
	go svc.asyncDispatcher.run()
	svc.wrapConnectionEventCallbacks()

	if config.Endpoint != nil {
		opts := []EndpointOpt{WithEndpointSubject(config.Endpoint.Subject)}
		if config.Endpoint.Metadata != nil {
			opts = append(opts, WithEndpointMetadata(config.Endpoint.Metadata))
		}
		if config.Endpoint.QueueGroup != "" {
			opts = append(opts, WithEndpointQueueGroup(config.Endpoint.QueueGroup))
		} else if config.QueueGroup != "" {
			opts = append(opts, WithEndpointQueueGroup(config.QueueGroup))
		}
		if err := svc.AddEndpoint("default", config.Endpoint.Handler, opts...); err != nil {
			svc.asyncDispatcher.close()
			return nil, err
		}
	}

	// Setup internal subscriptions.
	pingResponse := Ping{
		ServiceIdentity: svc.serviceIdentity(),
		Type:            PingResponseType,
	}

	handleVerb := func(verb Verb, valuef func() any) func(req Request) {
		return func(req Request) {
			response, _ := json.Marshal(valuef())
			if err := req.Respond(response); err != nil {
				if err := req.Error("500", fmt.Sprintf("Error handling %s request: %s", verb, err), nil); err != nil && config.ErrorHandler != nil {
					svc.asyncDispatcher.push(func() {
						config.ErrorHandler(svc, &NATSError{Subject: req.Subject(), Description: err.Error(), err: err})
					})
				}
			}
		}
	}

	for verb, source := range map[Verb]func() any{
		InfoVerb:  func() any { return svc.Info() },
		PingVerb:  func() any { return pingResponse },
		StatsVerb: func() any { return svc.Stats() },
	} {
		handler := handleVerb(verb, source)
		if err := svc.addVerbHandlers(nc, verb, handler); err != nil {
			svc.asyncDispatcher.close()
			return nil, err
		}
	}

	svc.started = time.Now().UTC()
	return svc, nil
}

func (s *service) AddEndpoint(name string, handler Handler, opts ...EndpointOpt) error {
	var options endpointOpts
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return err
		}
	}
	subject := name
	if options.subject != "" {
		subject = options.subject
	}
	queueGroup, noQueue := resolveQueueGroup(options.queueGroup, s.Config.QueueGroup, options.qgDisabled, s.Config.QueueGroupDisabled)
	return addEndpoint(s, name, subject, handler, options.metadata, queueGroup, noQueue, options.msgLimit, options.bytesLimit, options.middleware)
}

func addEndpoint(s *service, name, subject string, handler Handler, metadata map[string]string, queueGroup string, noQueue bool, msgLimit, bytesLimit int, endpointMW []Middleware) error {
	if !nameRegexp.MatchString(name) {
		return fmt.Errorf("%w: invalid endpoint name", ErrConfigValidation)
	}
	if !subjectRegexp.MatchString(subject) {
		return fmt.Errorf("%w: invalid endpoint subject", ErrConfigValidation)
	}
	if !subjectRegexp.MatchString(queueGroup) {
		return fmt.Errorf("%w: invalid endpoint queue group", ErrConfigValidation)
	}
	// Middleware order: service (outer) → endpoint → observer (inner, closest to handler).
	mws := append([]Middleware(nil), s.Config.Middleware...)
	mws = append(mws, endpointMW...)
	if s.Config.Observer != nil {
		mws = append(mws, ObserveEndpoint(s.Config.Observer, s.Config.Name, name, subject))
	}
	if len(mws) > 0 {
		handler = Chain(handler, mws...)
	}

	endpoint := &Endpoint{
		service: s,
		EndpointConfig: EndpointConfig{
			Subject:            subject,
			Handler:            handler,
			Metadata:           metadata,
			QueueGroup:         queueGroup,
			QueueGroupDisabled: noQueue,
		},
		Name: name,
	}

	var sub *nats.Subscription
	var err error
	var options = endpointOpts{
		msgLimit:   msgLimit,
		bytesLimit: bytesLimit,
	}

	handle := func(m *nats.Msg) {
		req := acquireRequest()
		req.msg = m
		req.ctx = s.baseCtx
		s.reqHandler(endpoint, req)
		releaseRequest(req)
	}
	if !noQueue {
		sub, err = s.nc.QueueSubscribe(subject, queueGroup, handle)
	} else {
		sub, err = s.nc.Subscribe(subject, handle)
	}
	if err != nil {
		return err
	}

	// Apply pending limits if configured
	if options.msgLimit != 0 || options.bytesLimit != 0 {
		if err := sub.SetPendingLimits(options.msgLimit, options.bytesLimit); err != nil {
			return err
		}
	}

	s.m.Lock()
	endpoint.subscription = sub
	s.endpoints = append(s.endpoints, endpoint)
	endpoint.stats = EndpointStats{
		Name:       name,
		Subject:    subject,
		QueueGroup: queueGroup,
	}
	s.m.Unlock()

	if s.Config.AdvertiseWhenPublic && isTruthy(metadata[MetaPublic]) {
		_ = AdvertiseRefresh(s.nc)
	}
	return nil
}

func (s *service) AddGroup(name string, opts ...GroupOpt) Group {
	var o groupOpts
	for _, opt := range opts {
		opt(&o)
	}
	queueGroup, noQueue := resolveQueueGroup(o.queueGroup, s.Config.QueueGroup, o.qgDisabled, s.Config.QueueGroupDisabled)
	return &group{
		service:            s,
		prefix:             name,
		queueGroup:         queueGroup,
		queueGroupDisabled: noQueue,
	}
}

// dispatch is responsible for calling any async callbacks
func (ac *asyncCallbacksHandler) run() {
	for {
		f, ok := <-ac.cbQueue
		if !ok || f == nil {
			return
		}
		f()
	}
}

// push enqueues an async callback without blocking forever if the queue is full.
func (ac *asyncCallbacksHandler) push(f func()) {
	if f == nil {
		return
	}
	select {
	case ac.cbQueue <- f:
	default:
		// drop under pressure; better than stalling request path
	}
}

func (ac *asyncCallbacksHandler) close() {
	if ac.closed {
		return
	}
	close(ac.cbQueue)
	ac.closed = true
}

func (c *Config) valid() error {
	if !nameRegexp.MatchString(c.Name) {
		return fmt.Errorf("%w: service name: name should not be empty and should consist of alphanumerical characters, dashes and underscores", ErrConfigValidation)
	}
	if !semVerRegexp.MatchString(c.Version) {
		return fmt.Errorf("%w: version: version should not be empty should match the SemVer format", ErrConfigValidation)
	}
	if c.QueueGroup != "" && !subjectRegexp.MatchString(c.QueueGroup) {
		return fmt.Errorf("%w: queue group: invalid queue group name", ErrConfigValidation)
	}

	return nil
}

func (s *service) wrapConnectionEventCallbacks() {
	s.m.Lock()
	defer s.m.Unlock()

	// Shared hub: multiple services on one Conn all receive events.
	s.unregClosed = onConnClosed(s.nc, func(c *nats.Conn) {
		_ = s.Stop()
	})
	s.unregError = onConnError(s.nc, func(c *nats.Conn, sub *nats.Subscription, err error) {
		if sub == nil {
			return
		}
		endpoint, match := s.matchSubscriptionSubject(sub.Subject)
		if !match {
			return
		}
		if s.Config.ErrorHandler != nil {
			s.Config.ErrorHandler(s, &NATSError{
				Subject:     sub.Subject,
				Description: err.Error(),
				err:         err,
			})
		}
		if endpoint != nil {
			endpoint.numErrors.Add(1)
			endpoint.lastError.Store(err.Error())
		}
		if s.Config.OnAsyncError == OnAsyncErrorStop {
			_ = s.Stop()
		}
	})
}

func (s *service) unwrapConnectionEventCallbacks() {
	if s.unregClosed != nil {
		s.unregClosed()
		s.unregClosed = nil
	}
	if s.unregError != nil {
		s.unregError()
		s.unregError = nil
	}
}

func (s *service) matchSubscriptionSubject(subj string) (*Endpoint, bool) {
	s.m.Lock()
	defer s.m.Unlock()
	for _, verbSub := range s.verbSubs {
		if verbSub.Subject == subj {
			return nil, true
		}
	}
	for _, e := range s.endpoints {
		if matchEndpointSubject(e.Subject, subj) {
			return e, true
		}
	}
	return nil, false
}

func matchEndpointSubject(endpointSubject, literalSubject string) bool {
	// Tokenize without allocating when subjects are equal (common async-error path).
	if endpointSubject == literalSubject {
		return true
	}
	ei, li := 0, 0
	for ei < len(endpointSubject) {
		// read endpoint token
		es := ei
		for ei < len(endpointSubject) && endpointSubject[ei] != '.' {
			ei++
		}
		et := endpointSubject[es:ei]
		if ei < len(endpointSubject) {
			ei++ // skip '.'
		}

		if et == ">" {
			// trailing ">" matches rest of subject (including empty remainder)
			return true
		}
		if li >= len(literalSubject) {
			return false
		}
		ls := li
		for li < len(literalSubject) && literalSubject[li] != '.' {
			li++
		}
		lt := literalSubject[ls:li]
		if li < len(literalSubject) {
			li++ // skip '.'
		}
		if et != "*" && et != lt {
			return false
		}
	}
	// Without a trailing ">", every subject token must be consumed; otherwise a
	// shorter endpoint would over-match a longer subject (e.g. "foo" vs "foo.bar").
	return li >= len(literalSubject)
}

// addVerbHandlers generates control handlers for a specific verb.
// Each request generates 3 subscriptions, one for the general verb
// affecting all services written with the framework, one that handles
// all services of a particular kind, and finally a specific service instance.
func (svc *service) addVerbHandlers(nc *nats.Conn, verb Verb, handler HandlerFunc) error {
	name := fmt.Sprintf("%s-all", verb.String())
	if err := svc.addInternalHandler(nc, verb, "", "", name, handler); err != nil {
		return err
	}
	name = fmt.Sprintf("%s-kind", verb.String())
	if err := svc.addInternalHandler(nc, verb, svc.Config.Name, "", name, handler); err != nil {
		return err
	}
	return svc.addInternalHandler(nc, verb, svc.Config.Name, svc.id, verb.String(), handler)
}

// addInternalHandler registers a control subject handler.
func (s *service) addInternalHandler(nc *nats.Conn, verb Verb, kind, id, name string, handler HandlerFunc) error {
	subj, err := ControlSubject(verb, kind, id)
	if err != nil {
		if stopErr := s.Stop(); stopErr != nil {
			return errors.Join(err, fmt.Errorf("stopping service: %w", stopErr))
		}
		return err
	}

	s.verbSubs[name], err = nc.Subscribe(subj, func(msg *nats.Msg) {
		handler(&request{msg: msg})
	})
	if err != nil {
		if stopErr := s.Stop(); stopErr != nil {
			return errors.Join(err, fmt.Errorf("stopping service: %w", stopErr))
		}
		return err
	}
	return nil
}

// reqHandler invokes the service request handler and modifies service stats.
// Endpoint counters are lock-free. stopped is atomic so we never take s.m on the
// hot path; double-check after inFlight.Add races cleanly with StopContext.
func (s *service) reqHandler(endpoint *Endpoint, req *request) {
	if s.stopped.Load() {
		return
	}
	s.inFlight.Add(1)
	if s.stopped.Load() {
		s.inFlight.Done()
		return
	}
	defer s.inFlight.Done()

	if req.ctx == nil {
		req.ctx = s.baseCtx
	}
	// Derive a per-request context canceled if the service stops mid-request.
	ctx, cancel := context.WithCancel(req.ctx)
	req.ctx = ctx
	defer cancel()

	start := time.Now()
	endpoint.Handler.Handle(req)
	endpoint.numRequests.Add(1)
	endpoint.processingNs.Add(time.Since(start).Nanoseconds())
	if req.respondError != nil {
		endpoint.numErrors.Add(1)
		endpoint.lastError.Store(req.respondError.Error())
	}
}

// Stop drains the endpoint subscriptions and marks the service as stopped.
func (s *service) Stop() error {
	return s.StopContext(context.Background())
}

// StopContext performs a graceful shutdown bounded by ctx.
func (s *service) StopContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.m.Lock()
	if s.stopped.Load() {
		s.m.Unlock()
		return nil
	}
	// Prevent new handlers from starting and cancel in-flight contexts.
	s.stopped.Store(true)
	if s.baseCancel != nil {
		s.baseCancel()
	}
	endpointsToStop := append(make([]*Endpoint, 0, len(s.endpoints)), s.endpoints...)
	verbSubs := make([]*nats.Subscription, 0, len(s.verbSubs))
	for _, sub := range s.verbSubs {
		verbSubs = append(verbSubs, sub)
	}
	s.endpoints = nil
	s.verbSubs = make(map[string]*nats.Subscription)
	doneHandler := s.DoneHandler
	s.m.Unlock()

	var drainErr error
	for _, e := range endpointsToStop {
		if err := e.stop(); err != nil && drainErr == nil {
			drainErr = err
		}
	}
	for _, sub := range verbSubs {
		if sub == nil {
			continue
		}
		if err := sub.Drain(); err != nil {
			if errors.Is(err, nats.ErrConnectionClosed) {
				break
			}
			if drainErr == nil {
				drainErr = fmt.Errorf("draining subscription for subject %q: %w", sub.Subject, err)
			}
		}
	}

	// Wait for in-flight handlers (or ctx timeout).
	done := make(chan struct{})
	go func() {
		s.inFlight.Wait()
		close(done)
	}()

	var waitErr error
	select {
	case <-done:
	case <-ctx.Done():
		waitErr = ctx.Err()
	}

	s.unwrapConnectionEventCallbacks()
	if doneHandler != nil {
		s.asyncDispatcher.push(func() { doneHandler(s) })
	}
	s.asyncDispatcher.close()

	return errors.Join(drainErr, waitErr)
}

func (s *service) serviceIdentity() ServiceIdentity {
	return ServiceIdentity{
		Name:     s.Config.Name,
		ID:       s.id,
		Version:  s.Config.Version,
		Metadata: s.Config.Metadata,
	}
}

// Info returns information about the service
func (s *service) Info() Info {
	s.m.Lock()
	defer s.m.Unlock()

	endpoints := make([]EndpointInfo, 0, len(s.endpoints))
	for _, e := range s.endpoints {
		endpoints = append(endpoints, EndpointInfo{
			Name:       e.Name,
			Subject:    e.Subject,
			QueueGroup: e.QueueGroup,
			Metadata:   e.Metadata,
		})
	}

	return Info{
		ServiceIdentity: s.serviceIdentity(),
		Type:            InfoResponseType,
		Description:     s.Config.Description,
		Endpoints:       endpoints,
	}
}

// Stats returns statistics for the service endpoint and all monitoring endpoints.
func (s *service) Stats() Stats {
	s.m.Lock()
	defer s.m.Unlock()

	stats := Stats{
		ServiceIdentity: s.serviceIdentity(),
		Endpoints:       make([]*EndpointStats, 0),
		Type:            StatsResponseType,
		Started:         s.started,
	}
	for _, endpoint := range s.endpoints {
		nreq := endpoint.numRequests.Load()
		proc := endpoint.processingNs.Load()
		var avg time.Duration
		if nreq > 0 {
			avg = time.Duration(proc / nreq)
		}
		lastErr, _ := endpoint.lastError.Load().(string)
		endpointStats := &EndpointStats{
			Name:                  endpoint.stats.Name,
			Subject:               endpoint.stats.Subject,
			QueueGroup:            endpoint.stats.QueueGroup,
			NumRequests:           int(nreq),
			NumErrors:             int(endpoint.numErrors.Load()),
			LastError:             lastErr,
			ProcessingTime:        time.Duration(proc),
			AverageProcessingTime: avg,
		}
		if s.StatsHandler != nil {
			data, _ := json.Marshal(s.StatsHandler(endpoint))
			endpointStats.Data = data
		}
		stats.Endpoints = append(stats.Endpoints, endpointStats)
	}
	return stats
}

// Reset resets all statistics on a service instance.
func (s *service) Reset() {
	s.m.Lock()
	for _, endpoint := range s.endpoints {
		endpoint.reset()
	}
	s.started = time.Now().UTC()
	s.m.Unlock()
}

// Stopped informs whether [Stop] was executed on the service.
func (s *service) Stopped() bool {
	return s.stopped.Load()
}
