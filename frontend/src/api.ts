import type { Project, Manifest, CreateProjectRequest, ReviewRequest } from './types';

const API_BASE = '/api';

class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = 'ApiError';
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.text().catch(() => 'Unknown error');
    throw new ApiError(body || `Request failed: ${res.status}`, res.status);
  }

  return res.json();
}

export async function createProject(data: CreateProjectRequest): Promise<Project> {
  return request<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getProject(projectId: string): Promise<Project> {
  return request<Project>(`/projects/${projectId}`);
}

export async function getManifest(projectId: string): Promise<Manifest> {
  return request<Manifest>(`/projects/${projectId}/manifest`);
}

export async function approveProject(projectId: string, data: ReviewRequest): Promise<{ status: string }> {
  return request(`/projects/${projectId}/approve`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function rejectProject(projectId: string, data: ReviewRequest): Promise<{ status: string }> {
  return request(`/projects/${projectId}/reject`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function retryProject(projectId: string): Promise<{ status: string }> {
  return request(`/projects/${projectId}/retry`, {
    method: 'POST',
  });
}

export function getDownloadUrl(projectId: string): string {
  return `${API_BASE}/projects/${projectId}/download`;
}

export async function downloadVideo(projectId: string): Promise<void> {
  const response = await fetch(getDownloadUrl(projectId));
  if (!response.ok) throw new ApiError('Download failed', response.status);
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `video_${projectId}.mp4`;
  a.click();
  window.URL.revokeObjectURL(url);
}

export async function healthCheck(): Promise<{ status: string }> {
  return request('/health');
}

export { ApiError };
