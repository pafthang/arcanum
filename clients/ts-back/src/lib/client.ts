import { HttpClient, type ClientOptions } from './http.js';
import { authResource } from './resources/auth.js';
import { bandwidthResource } from './resources/bandwidth.js';
import { configProfilesResource } from './resources/config-profiles.js';
import { connectionsResource } from './resources/connections.js';
import { hostsResource } from './resources/hosts.js';
import { hwidResource } from './resources/hwid.js';
import { infraResource } from './resources/infra.js';
import { integrationsResource } from './resources/integrations.js';
import { metadataResource } from './resources/metadata.js';
import { nodesResource } from './resources/nodes.js';
import { passkeysResource } from './resources/passkeys.js';
import { pluginsResource } from './resources/plugins.js';
import { settingsResource } from './resources/settings.js';
import { snippetsResource } from './resources/snippets.js';
import { squadsResource } from './resources/squads.js';
import { subscriptionResource } from './resources/subscription.js';
import { systemResource } from './resources/system.js';
import { tokensResource } from './resources/tokens.js';
import { usersResource } from './resources/users.js';
import { session } from './session.svelte.js';

export type RemnawaveClient = ReturnType<typeof createRemnawave>;

export function createRemnawave(opts: ClientOptions = {}) {
	const http = new HttpClient({
		...opts,
		getToken: opts.getToken ?? (() => session.token),
		onUnauthorized: opts.onUnauthorized ?? (() => session.clear())
	});

	return {
		http,
		session,
		auth: authResource(http),
		users: usersResource(http),
		nodes: nodesResource(http),
		hosts: hostsResource(http),
		system: systemResource(http),
		configProfiles: configProfilesResource(http),
		squads: squadsResource(http),
		snippets: snippetsResource(http),
		settings: settingsResource(http),
		hwid: hwidResource(http),
		tokens: tokensResource(http),
		passkeys: passkeysResource(http),
		plugins: pluginsResource(http),
		metadata: metadataResource(http),
		subscription: subscriptionResource(http),
		connections: connectionsResource(http),
		integrations: integrationsResource(http),
		infra: infraResource(http),
		bandwidth: bandwidthResource(http)
	};
}
