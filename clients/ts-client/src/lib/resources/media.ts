import type { HttpClient } from '../http.js';
import { session } from '../session.svelte.js';
import type { BlobInfo } from '../types.js';

export function mediaResource(http: HttpClient) {
	const resolveSpace = (sid?: string) => sid || session.spaceId;

	return {
		async listBlobs(spaceId?: string): Promise<BlobInfo[]> {
			const sid = resolveSpace(spaceId);
			return http.get<BlobInfo[]>(`/api/spaces/${sid}/blobs`);
		},
		async uploadBlob(file: File, spaceId?: string): Promise<BlobInfo> {
			const sid = resolveSpace(spaceId);
			const formData = new FormData();
			formData.append('file', file);
			return http.post<BlobInfo>(`/api/spaces/${sid}/blobs`, formData);
		},
		async getBlob(blobId: string, spaceId?: string): Promise<BlobInfo> {
			const sid = resolveSpace(spaceId);
			return http.get<BlobInfo>(`/api/spaces/${sid}/blobs/${blobId}`);
		},
		async deleteBlob(blobId: string, spaceId?: string): Promise<{ ok: boolean }> {
			const sid = resolveSpace(spaceId);
			return http.delete<{ ok: boolean }>(`/api/spaces/${sid}/blobs/${blobId}`);
		},
		async getSignedURL(blobId: string, spaceId?: string): Promise<{ url: string }> {
			const sid = resolveSpace(spaceId);
			return http.get<{ url: string }>(`/api/spaces/${sid}/blobs/${blobId}/url`);
		},
		getBlobURL(blobId: string, spaceId?: string): string {
			const sid = resolveSpace(spaceId);
			return `/api/spaces/${sid}/blobs/${blobId}`;
		}
	};
}
