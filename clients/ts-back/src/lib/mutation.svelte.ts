export type MutationStatus = 'idle' | 'pending' | 'success' | 'error';

export type MutationOptions<T, V> = {
	fn: (variables: V) => Promise<T>;
	onSuccess?: (data: T, variables: V) => void;
	onError?: (error: Error, variables: V) => void;
};

export type Mutation<T, V> = {
	readonly data: T | undefined;
	readonly error: Error | undefined;
	readonly status: MutationStatus;
	readonly pending: boolean;
	mutate: (variables: V) => Promise<T>;
	reset: () => void;
};

export function createMutation<T, V = void>(opts: MutationOptions<T, V>): Mutation<T, V> {
	let data = $state.raw<T | undefined>(undefined);
	let error = $state<Error | undefined>(undefined);
	let status = $state<MutationStatus>('idle');

	async function mutate(variables: V): Promise<T> {
		status = 'pending';
		error = undefined;
		try {
			const result = await opts.fn(variables);
			data = result;
			status = 'success';
			opts.onSuccess?.(result, variables);
			return result;
		} catch (e) {
			const err = e instanceof Error ? e : new Error(String(e));
			error = err;
			status = 'error';
			opts.onError?.(err, variables);
			throw err;
		}
	}

	function reset() {
		data = undefined;
		error = undefined;
		status = 'idle';
	}

	return {
		get data() {
			return data;
		},
		get error() {
			return error;
		},
		get status() {
			return status;
		},
		get pending() {
			return status === 'pending';
		},
		mutate,
		reset
	};
}
