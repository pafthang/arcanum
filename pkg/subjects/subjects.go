// Package subjects defines the NATS subject map for the micro platform.
package subjects

const (
	// Public namespace (gate → services). Prefer mini.WithPublicSubject.
	PublicPrefix = "public"

	// Internal request/reply (service → service).
	InternalPrefix = "internal"

	// Domain events (JetStream / core NATS pub-sub).
	EventsPrefix = "events"

	// Commands for workers.
	CommandsPrefix = "commands"
)

// Public HTTP-facing subjects (also advertised via mini public metadata).
const (
	// logs
	PublicLogsList  = "public.logs.list"
	PublicLogsGet   = "public.logs.get"
	PublicLogsStats = "public.logs.stats"

	// health
	PublicHealth = "public.health.check"

	// space
	PublicSpaceAuthLogin       = "public.space.auth.login"
	PublicSpaceAuthRegister    = "public.space.auth.register"
	PublicSpaceAuthSwitchSpace = "public.space.auth.switch_space"
	PublicSpaceSpacesList      = "public.space.spaces.list"
	PublicSpaceSpacesCreate    = "public.space.spaces.create"
	PublicSpaceSpacesGet       = "public.space.spaces.get"
	PublicSpaceMembersList     = "public.space.members.list"
	PublicSpaceMembersInvite   = "public.space.members.invite"
	PublicSpaceMembersUpdate   = "public.space.members.update"
	PublicSpaceTeamsList       = "public.space.teams.list"
	PublicSpaceTeamsCreate     = "public.space.teams.create"
	PublicSpaceTeamsGet        = "public.space.teams.get"
	PublicSpaceTeamMembersAdd  = "public.space.teams.members.add"
	PublicSpaceKeysList        = "public.space.keys.list"
	PublicSpaceKeysCreate      = "public.space.keys.create"

	// work
	PublicWorkIssueList           = "public.work.issue.list"
	PublicWorkIssueCreate         = "public.work.issue.create"
	PublicWorkIssueGet            = "public.work.issue.get"
	PublicWorkIssueUpdate         = "public.work.issue.update"
	PublicWorkCommentList         = "public.work.comment.list"
	PublicWorkCommentCreate       = "public.work.comment.create"
	PublicWorkOverview            = "public.work.overview"
	PublicWorkLabelList           = "public.work.label.list"
	PublicWorkLabelCreate         = "public.work.label.create"
	PublicWorkIssueRelationCreate = "public.work.issue.relation.create"

	// comms
	PublicCommsChannelList   = "public.comms.channel.list"
	PublicCommsChannelCreate = "public.comms.channel.create"
	PublicCommsChannelGet    = "public.comms.channel.get"
	PublicCommsMessageList   = "public.comms.message.list"
	PublicCommsMessageCreate = "public.comms.message.create"
	PublicCommsChannelWS     = "public.comms.channel.ws"

	// agents
	PublicAgentsRunList     = "public.agents.run.list"
	PublicAgentsRunCreate   = "public.agents.run.create"
	PublicAgentsRunGet      = "public.agents.run.get"
	PublicAgentsRunCancel   = "public.agents.run.cancel"
	PublicAgentsSessionGet  = "public.agents.session.get"
	PublicAgentsMemoryList  = "public.agents.memory.list"
	PublicAgentsMemoryPut   = "public.agents.memory.put"
	PublicAgentsSkillList   = "public.agents.skill.list"
	PublicAgentsSkillCreate = "public.agents.skill.create"

	// integ
	PublicIntegConnectorList   = "public.integ.connector.list"
	PublicIntegConnectorCreate = "public.integ.connector.create"
	PublicIntegConnectorGet    = "public.integ.connector.get"
	PublicIntegConnectorUpdate = "public.integ.connector.update"
	PublicIntegRepoList        = "public.integ.repo.list"
	PublicIntegRepoCreate      = "public.integ.repo.create"
	PublicIntegWebhookList     = "public.integ.webhook.list"
	PublicIntegWebhookCreate   = "public.integ.webhook.create"
	PublicIntegDeliveryList    = "public.integ.delivery.list"
	PublicIntegHookIngest      = "public.integ.hook.ingest"
	PublicIntegHookGitHub      = "public.integ.hook.github"

	// media
	PublicMediaBlobList    = "public.media.blob.list"
	PublicMediaBlobCreate  = "public.media.blob.create"
	PublicMediaBlobGet     = "public.media.blob.get"
	PublicMediaBlobContent = "public.media.blob.content"
	PublicMediaBlobURL     = "public.media.blob.url"
	PublicMediaBlobDelete  = "public.media.blob.delete"
)

// Internal RPC.
const (
	InternalIntegConnectorGet  = "internal.integ.connector.get"
	InternalIntegConnectorList = "internal.integ.connector.list"
	InternalIntegIngest        = "internal.integ.ingest"
	InternalAgentsRunGet       = "internal.agents.run.get"
	// logg: activity (durable audit events)
	InternalActivityAppend     = "internal.logg.activity.append"
	InternalActivityList       = "internal.logg.activity.list"
	InternalActivityListTarget = "internal.logg.activity.list_target"

	// space
	InternalSpaceGet         = "internal.space.get"
	InternalSpaceListForUser = "internal.space.list_for_user"
	InternalSpaceUserGet     = "internal.space.user.get"
	InternalSpaceCan         = "internal.space.can"

	// work
	InternalWorkIssueGet  = "internal.work.issue.get"
	InternalWorkIssueList = "internal.work.issue.list"
	InternalWorkOverview  = "internal.work.overview"

	// comms
	InternalCommsChannelGet    = "internal.comms.channel.get"
	InternalCommsChannelList   = "internal.comms.channel.list"
	InternalCommsMessageCreate = "internal.comms.message.create"
	InternalCommsInbound       = "internal.comms.inbound"

	// media
	InternalMediaGet      = "internal.media.get"
	InternalMediaGetBytes = "internal.media.get_bytes"
)

// Events (publish).
const (
	EventAccessRequest     = "events.access.request"
	EventWorkIssueCreated  = "events.work.issue.created"
	EventWorkIssueUpdated  = "events.work.issue.updated"
	EventWorkIssueAssigned = "events.work.issue.assigned"

	EventCommsChannelCreated   = "events.comms.channel.created"
	EventCommsMessageCreated   = "events.comms.message.created"
	EventCommsChannelWSPattern = "events.comms.{spaceId}.{channelId}"
	EventIntegMessageInbound   = "events.integ.message.inbound"
	EventIntegInbound          = "events.integ.inbound"
	EventIntegGitHubActivity   = "events.integ.github.activity"
	EventIntegChannelMessage   = "events.integ.channel.message"
	EventIntegWebhookDelivered = "events.integ.webhook.delivered"
	EventAgentsRunStarted      = "events.agents.run.started"
	EventAgentsRunFinished     = "events.agents.run.finished"
)

// Commands.
const (
	CommandAgentsRunStart  = "commands.agents.run.start"
	CommandAgentsRunCancel = "commands.agents.run.cancel"

	// Platform lifecycle (ops → all services).
	// Payload: events.Envelope with LifecycleCommand data.
	CommandPlatformReload  = "commands.platform.reload"  // soft: reopen DBs
	CommandPlatformRestart = "commands.platform.restart" // hard: process exit (supervisor restarts)
)

// EventCommsChannel is the per-channel fan-out subject used by gate WS.
func EventCommsChannel(spaceID, channelID string) string {
	return "events.comms." + spaceID + "." + channelID
}
