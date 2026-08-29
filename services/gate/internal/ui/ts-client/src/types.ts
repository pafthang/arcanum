export interface Issue {
  id: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'done';
  space_id: string;
  created_by: string;
  assigned_to?: string;
  due_date?: string;
  project_id?: string;
  cycle_id?: string;
  labels?: string[];
  created_at: string;
  updated_at: string;
}

export interface CreateIssueInput {
  title: string;
  description?: string;
  project_id?: string;
  cycle_id?: string;
  labels?: string[];
}
