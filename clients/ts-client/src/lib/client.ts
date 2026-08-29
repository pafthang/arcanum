import { HttpClient, type ClientOptions } from './http.js';
import { authResource } from './resources/auth.js';
import { spaceResource } from './resources/space.js';
import { teamsResource } from './resources/teams.js';
import { workResource } from './resources/work.js';
import { agentsResource } from './resources/agents.js';
import { commsResource } from './resources/comms.js';
import { integResource } from './resources/integ.js';
import { mediaResource } from './resources/media.js';
import { runtimeResource } from './resources/runtime.js';
import { loggResource } from './resources/logg.js';
import { session } from './session.svelte.js';

export type ArcanumClient = ReturnType<typeof createArcanumClient>;

export function createArcanumClient(opts: ClientOptions = {}) {
	const http = new HttpClient({
		...opts,
		getToken: opts.getToken ?? (() => session.token),
		onUnauthorized: opts.onUnauthorized ?? (() => session.clear())
	});

	return {
		http,
		session,
		auth: authResource(http),
		space: spaceResource(http),
		teams: teamsResource(http),
		work: workResource(http),
		agents: agentsResource(http),
		comms: commsResource(http),
		integ: integResource(http),
		media: mediaResource(http),
		runtime: runtimeResource(http),
		logg: loggResource(http)
	};
}
