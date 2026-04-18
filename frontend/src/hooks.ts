import { useState, useEffect, useCallback, useRef } from 'react';
import { getProject, getManifest } from './api';
import { getProjectIds, removeProjectId } from './storage';
import { isProcessing, type Project, type Manifest } from './types';

// Fetch a single project with auto-polling when processing
export function useProject(projectId: string | undefined) {
  const [project, setProject] = useState<Project | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchProject = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await getProject(projectId);
      setProject(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch project');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    setLoading(true);
    fetchProject();
  }, [fetchProject]);

  // Poll every 5s when processing
  useEffect(() => {
    if (!project || !isProcessing(project.status)) return;
    const interval = setInterval(fetchProject, 5000);
    return () => clearInterval(interval);
  }, [project?.status, fetchProject]);

  return { project, loading, error, refetch: fetchProject };
}

// Fetch manifest
export function useManifest(projectId: string | undefined) {
  const [manifest, setManifest] = useState<Manifest | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchManifest = useCallback(async () => {
    if (!projectId) return;
    try {
      const data = await getManifest(projectId);
      setManifest(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch manifest');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    fetchManifest();
  }, [fetchManifest]);

  return { manifest, loading, error, refetch: fetchManifest };
}

// Fetch all projects from localStorage
export function useProjects() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const mountedRef = useRef(true);

  const fetchAll = useCallback(async () => {
    const ids = getProjectIds();
    if (ids.length === 0) {
      setProjects([]);
      setLoading(false);
      return;
    }

    const results: Project[] = [];
    for (const id of ids) {
      try {
        const project = await getProject(id);
        if (mountedRef.current) results.push(project);
      } catch {
        // If 404, remove from localStorage
        removeProjectId(id);
      }
    }

    if (mountedRef.current) {
      setProjects(results);
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    mountedRef.current = true;
    fetchAll();
    return () => { mountedRef.current = false; };
  }, [fetchAll]);

  // Auto-refresh every 5s if any project is processing
  useEffect(() => {
    const hasProcessing = projects.some((p) => isProcessing(p.status));
    if (!hasProcessing) return;
    const interval = setInterval(fetchAll, 5000);
    return () => clearInterval(interval);
  }, [projects, fetchAll]);

  return { projects, loading, refetch: fetchAll };
}
