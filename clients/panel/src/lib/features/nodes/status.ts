export type NodeLike = {
	address: string;
	port: number | null;
	isConnected: boolean;
	isConnecting: boolean;
	isDisabled: boolean;
	configProfile?: unknown;
};

export function nodeStatus(node: NodeLike): string {
	if (node.isDisabled) return 'DISABLED';
	if (node.isConnecting) return 'CONNECTING';
	if (node.isConnected) return 'CONNECTED';
	return 'OFFLINE';
}

export function nodeEndpoint(node: Pick<NodeLike, 'address' | 'port'>): string {
	return node.port ? `${node.address}:${node.port}` : node.address;
}

export type NodeProfile = {
	uuid?: string;
	inbounds: { uuid: string; tag?: string }[];
};

export function nodeProfile(node: Pick<NodeLike, 'configProfile'>): NodeProfile {
	const raw = node.configProfile as
		| { activeConfigProfileUuid?: string | null; activeInbounds?: { uuid: string; tag?: string }[] }
		| null
		| undefined;
	return {
		uuid: raw?.activeConfigProfileUuid ?? undefined,
		inbounds: raw?.activeInbounds ?? []
	};
}
