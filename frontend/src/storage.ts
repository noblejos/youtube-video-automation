const STORAGE_KEY = 'yt_automation_projects';

export function getProjectIds(): string[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    return JSON.parse(raw) as string[];
  } catch {
    return [];
  }
}

export function saveProjectId(id: string): void {
  const ids = getProjectIds();
  if (!ids.includes(id)) {
    ids.unshift(id); // newest first
    localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
  }
}

export function removeProjectId(id: string): void {
  const ids = getProjectIds().filter((pid) => pid !== id);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
}
