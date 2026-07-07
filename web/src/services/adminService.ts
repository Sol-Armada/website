import { requestJson } from '@/services/http'

export interface AdminOverviewData {
  totalMembers: number
  totalEvents: number
  totalTokens: number
  activeThisMonth: number
  uniqueAttendees: number
  averageAttendance: number
}

export interface AttendanceRecord {
  id: string
  name: string
  submittedBy: string
  participantCount: number
  recorded: boolean
  successful: boolean
  dateCreated: string
  awardTokens: boolean
}

export interface CreateAttendanceRecordRequest {
  submittedBy: string | null
  name: string
  participantIds: string[]
  managerIds: string[]
  awardTokens?: boolean
}

export interface AttendanceNameMutationRequest {
  name: string
}

export interface TokenTransaction {
  id: string
  memberId: string
  memberName: string
  amount: number
  reason: string
  createdAt: string
  comment?: string
  giverId?: string
  attendanceId?: string
  attendanceName?: string
}

export interface TokenPeriodAnalytics {
  windowStart: string
  windowEnd: string
  totalEarnings: number
  totalSpending: number
  netAmount: number
  averageEarningPerMember: number
  averageSpendingPerMember: number
  averageEarningPerTransaction: number
  averageSpendingPerTransaction: number
  earningTransactionCount: number
  spendingTransactionCount: number
  earningMemberCount: number
  spendingMemberCount: number
}

export interface TokenReasonAggregation {
  reason: string
  transactionCount: number
  netAmount: number
  totalEarnings: number
  totalSpending: number
}

export interface TokenLedgerAnalytics {
  week: TokenPeriodAnalytics
  month: TokenPeriodAnalytics
  reasons: TokenReasonAggregation[]
}

export interface AttendanceAnalytics {
  windowStart: string
  windowEnd: string
  uniqueAttendeesLast30Days: number
  inactiveMembersLast30Days: number
  mostPopularEventLast30Days: string
  mostPopularEventAttendanceLast30Days: number
  totalEventsLast30Days: number
}

export interface MemberSummary {
  id: string
  username: string
  rank: string
  attendance: number
  tokenBalance: number
  rsiHandle?: string
  profileImage?: string
  onTime?: boolean
  isManager?: boolean
}

export interface AttendanceEditPayload {
  record: AttendanceRecord
  participants: MemberSummary[]
}

export interface UpdateAttendanceRecordRequest {
  name: string
  recorded: boolean
  successful: boolean
  awardTokens: boolean
  participantIds: string[]
  onTimeParticipantIds: string[]
}

interface PaginatedResponse<T> {
  records?: T[]
  members?: T[]
  page: number
  limit: number
}

export const projectStatuses = new Map<number, string>()
projectStatuses.set(1, 'Not Started')
projectStatuses.set(2, 'In Progress')
projectStatuses.set(3, 'Completed')
projectStatuses.set(4, 'On Hold')
projectStatuses.set(5, 'Cancelled')

export interface ProjectSummary {
  id: string
  name: string
  description: string
  statusId: number
  statusName: string
  ownerId?: string | null
  ownerName?: string | null
  dueAt?: string | null
  progress: number
  totalTasks: number
  doneTasks: number
  memberCount: number
  createdAt: string
  updatedAt: string
}

export interface ProjectListResponse {
  projects: ProjectSummary[]
}

type RawProjectOwner = {
  id?: string | null
  name?: string | null
}

type RawProjectTaskSummary = {
  assignee?: { id?: string | null } | null
  status?: { name?: string | null, isDoneState?: boolean | null } | null
}

type RawProject = {
  id?: string
  Id?: string
  projectId?: string
  projectID?: string
  _id?: string
  uuid?: string
  projectUuid?: string
  projectUUID?: string
  name: string
  description: string
  statusId?: number
  statusName?: string
  status?: number | string | { id?: number, name?: string } | null
  ownerId?: string | null
  ownerName?: string | null
  owner?: RawProjectOwner | null
  dueAt?: string | null
  tasks?: RawProjectTaskSummary[] | null
  totalTasks?: number
  doneTasks?: number
  memberCount?: number
  progress?: number
  createdAt?: string
  updatedAt?: string
}

type ListProjectsRawResponse = RawProject[] | { projects?: RawProject[] } | null | undefined

function looksDone(task: RawProjectTaskSummary): boolean {
  if (task.status?.isDoneState) {
    return true
  }
  const statusName = (task.status?.name || '').toLowerCase()
  return statusName.includes('done') || statusName.includes('complete')
}

function statusIdFromRaw(project: RawProject): number {
  if (typeof project.statusId === 'number' && Number.isFinite(project.statusId)) {
    return project.statusId
  }

  if (typeof project.status === 'number' && Number.isFinite(project.status)) {
    return project.status
  }

  if (typeof project.status === 'object' && project.status && typeof project.status.id === 'number') {
    return project.status.id
  }

  if (typeof project.status === 'string') {
    const parsed = Number(project.status)
    if (Number.isFinite(parsed)) {
      return parsed
    }

    const lowered = project.status.toLowerCase()
    for (const [id, name] of projectStatuses.entries()) {
      if (name.toLowerCase() === lowered) {
        return id
      }
    }
  }

  return 1
}

function statusNameFromRaw(project: RawProject, statusId: number): string {
  if (project.statusName && project.statusName.trim()) {
    return project.statusName
  }

  if (typeof project.status === 'string' && project.status.trim()) {
    const parsed = Number(project.status)
    if (!Number.isFinite(parsed)) {
      return project.status
    }
  }

  if (typeof project.status === 'object' && project.status && typeof project.status.name === 'string' && project.status.name.trim()) {
    return project.status.name
  }

  return projectStatuses.get(statusId) || 'Not Started'
}

function normalizeDueAt(dueAt?: string | null): string | null {
  if (!dueAt) {
    return null
  }
  if (dueAt.startsWith('0001-01-01')) {
    return null
  }
  return dueAt
}

function firstNonEmptyString(...values: Array<string | null | undefined>): string {
  for (const value of values) {
    if (typeof value === 'string' && value.trim()) {
      return value
    }
  }
  return ''
}

function resolveProjectId(project: RawProject): string {
  return firstNonEmptyString(
    project.id,
    project.Id,
    project.projectId,
    project.projectID,
    project._id,
    project.uuid,
    project.projectUuid,
    project.projectUUID,
  )
}

function normalizeProjectSummary(project: RawProject, fallbackId?: string): ProjectSummary {
  const tasks = Array.isArray(project.tasks) ? project.tasks : []
  const totalTasks = typeof project.totalTasks === 'number' ? project.totalTasks : tasks.length
  const doneTasks = typeof project.doneTasks === 'number'
    ? project.doneTasks
    : tasks.filter(task => looksDone(task)).length

  const memberCount = typeof project.memberCount === 'number'
    ? project.memberCount
    : new Set(tasks
      .map(task => task.assignee?.id)
      .filter(Boolean))
      .size

  const progress = typeof project.progress === 'number'
    ? project.progress
    : (totalTasks > 0 ? Math.round((doneTasks * 100) / totalTasks) : 0)

  const statusId = statusIdFromRaw(project)
  const ownerId = project.ownerId ?? project.owner?.id ?? null
  const ownerName = project.ownerName ?? project.owner?.name ?? null
  const id = resolveProjectId(project) || fallbackId || firstNonEmptyString(project.name, project.createdAt, project.updatedAt)

  return {
    id,
    name: project.name,
    description: project.description,
    statusId,
    statusName: statusNameFromRaw(project, statusId),
    ownerId,
    ownerName,
    dueAt: normalizeDueAt(project.dueAt),
    progress,
    totalTasks,
    doneTasks,
    memberCount,
    createdAt: project.createdAt || '',
    updatedAt: project.updatedAt || '',
  }
}

function normalizeProjectListResponse(response: ListProjectsRawResponse): ProjectListResponse {
  if (!response) {
    return { projects: [] }
  }
  if (Array.isArray(response)) {
    return { projects: response.map((project, index) => normalizeProjectSummary(project, `project-${index + 1}`)) }
  }
  return { projects: (response.projects ?? []).map((project, index) => normalizeProjectSummary(project, `project-${index + 1}`)) }
}

export interface CreateProjectRequest {
  name: string
  description: string
  statusId: number
  ownerId?: string | null
  dueAt?: string | null
}

export interface ProjectStatus {
  id: number
  name: string
}

export type TaskPriority = 0 | 1 | 2
export type KanbanStatus = 'backlog' | 'todo' | 'inprogress' | 'review' | 'done' | string

export interface ProjectTaskStatus {
  projectId: string
  name: string
  position: number
  color: string
}

export interface ProjectTaskAssignee {
  id: string
  name?: string
}

export interface ProjectTaskActivity {
  id: string
  summary: string
  time?: string
}

export interface ProjectTask {
  id: string
  projectId: string
  title: string
  description: string
  position: number
  priority: TaskPriority
  assignee?: ProjectTaskAssignee | null
  dueAt?: string | null
  status?: ProjectTaskStatus | null
  parentTask?: { id: string, title?: string } | null
  activity?: ProjectTaskActivity[]
  createdAt: string
  updatedAt: string
}

export interface ProjectTaskListResponse {
  tasks: ProjectTask[]
}

export interface CreateProjectTaskRequest {
  title: string
  description: string
  priority: number
  assignee: string
  dueAt?: string | null
  status: string
  parentTaskId?: string | null
}

export interface UpdateProjectTaskRequest {
  title: string
  description: string
  priority: number
  assignee: string
  dueAt?: string | null
  status: string
  parentTaskId?: string | null
}

type RawProjectTaskActivity = {
  id?: string
  summary?: string
  actorInitials?: string
  actorName?: string
  action?: string
  time?: string
  details?: string
}

type RawProjectTask = {
  id: string
  projectId: string
  title: string
  description: string
  position: number
  priority: number
  assignee?: ProjectTaskAssignee | null
  dueAt?: string | null
  status?: ProjectTaskStatus | null
  parentTask?: { id: string, title?: string } | null
  activity?: RawProjectTaskActivity[]
  createdAt: string
  updatedAt: string
}

function normalizePriority(priority: number): TaskPriority {
  if (priority >= 2) {
    return 2
  }
  if (priority <= 0) {
    return 0
  }
  return 1
}

function normalizeTaskActivity(activity: RawProjectTaskActivity, index: number): ProjectTaskActivity {
  const summary = (activity.summary || '').trim()
    || (activity.details || '').trim()
    || [activity.actorName, activity.action].filter(Boolean).join(' ').trim()

  return {
    id: activity.id || `activity-${index + 1}`,
    summary,
    time: activity.time,
  }
}

function normalizeTask(task: RawProjectTask): ProjectTask {
  const normalizedActivity = (task.activity ?? [])
    .map((entry, index) => normalizeTaskActivity(entry, index))
    .filter(entry => entry.summary)

  return {
    ...task,
    priority: normalizePriority(task.priority),
    dueAt: task.dueAt ?? null,
    assignee: task.assignee ?? null,
    status: task.status ?? null,
    parentTask: task.parentTask ?? null,
    activity: normalizedActivity,
  }
}

function normalizeTaskListResponse(response: { tasks?: RawProjectTask[] }): ProjectTaskListResponse {
  return {
    tasks: (response.tasks ?? []).map(task => normalizeTask(task)),
  }
}

export const adminService = {
  async getOverview(): Promise<AdminOverviewData> {
    return requestJson<AdminOverviewData>('/api/admin/overview')
  },

  async getAttendance(limit = 50, page = 1, search?: string): Promise<PaginatedResponse<AttendanceRecord>> {
    return requestJson<PaginatedResponse<AttendanceRecord>>('/api/admin/attendance', undefined, {
      limit,
      page,
      search,
    })
  },

  async getAttendanceRecord(id: string): Promise<AttendanceRecord> {
    return requestJson<AttendanceRecord>(`/api/admin/attendance/${id}`)
  },

  async getAttendanceEditPayload(id: string): Promise<AttendanceEditPayload> {
    return requestJson<AttendanceEditPayload>(`/api/admin/attendance/${id}/edit`)
  },

  async getAvailableAttendanceNames(): Promise<string[]> {
    return requestJson<string[]>('/api/admin/attendance/names')
  },

  async createAttendanceName(payload: AttendanceNameMutationRequest): Promise<{ name: string }> {
    return requestJson<{ name: string }>('/api/admin/attendance/names', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async deleteAttendanceName(payload: AttendanceNameMutationRequest): Promise<{ name: string }> {
    return requestJson<{ name: string }>('/api/admin/attendance/names', {
      method: 'DELETE',
      body: JSON.stringify(payload),
    })
  },

  async createAttendanceRecord(payload: CreateAttendanceRecordRequest): Promise<AttendanceRecord> {
    return requestJson<AttendanceRecord>('/api/admin/attendance', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async getTokenLedger(limit = 50, page = 1, search?: string): Promise<PaginatedResponse<TokenTransaction>> {
    return requestJson<PaginatedResponse<TokenTransaction>>('/api/admin/token-ledger', undefined, {
      limit,
      page,
      search,
    })
  },

  async getTokenLedgerAnalytics(): Promise<TokenLedgerAnalytics> {
    return requestJson<TokenLedgerAnalytics>('/api/admin/token-ledger/analytics')
  },

  async getAttendanceAnalytics(): Promise<AttendanceAnalytics> {
    return requestJson<AttendanceAnalytics>('/api/admin/attendance/analytics')
  },

  async getMembers(limit = 50, page = 1, search?: string): Promise<PaginatedResponse<MemberSummary>> {
    return requestJson<PaginatedResponse<MemberSummary>>('/api/admin/members', undefined, {
      limit,
      page,
      search,
    })
  },

  async getMembersByAttendance(attendanceId: string): Promise<MemberSummary[]> {
    return requestJson<MemberSummary[]>(`/api/admin/attendance/${attendanceId}/members`)
  },

  async updateAttendanceRecord(attendanceId: string, payload: UpdateAttendanceRecordRequest): Promise<AttendanceEditPayload> {
    return requestJson<AttendanceEditPayload>(`/api/admin/attendance/${attendanceId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    })
  },

  async listProjects(): Promise<ProjectListResponse> {
    const response = await requestJson<ListProjectsRawResponse>('/api/admin/projects')
    return normalizeProjectListResponse(response)
  },

  async createProject(request: CreateProjectRequest): Promise<ProjectSummary> {
    const response = await requestJson<RawProject | null>('/api/admin/projects', {
      method: 'POST',
      body: JSON.stringify(request),
    })
    if (!response) {
      throw new Error('Unexpected empty project response')
    }
    return normalizeProjectSummary(response)
  },

  async listProjectStatuses(): Promise<ProjectStatus[]> {
    return requestJson<ProjectStatus[]>('/api/admin/project-statuses')
  },

  async listProjectTaskStatuses(projectId: string): Promise<ProjectTaskStatus[]> {
    return requestJson<ProjectTaskStatus[]>(`/api/admin/projects/${encodeURIComponent(projectId)}/task-statuses`)
  },

  async listProjectTasks(projectId: string): Promise<ProjectTaskListResponse> {
    const response = await requestJson<{ tasks?: RawProjectTask[] }>(`/api/admin/projects/${encodeURIComponent(projectId)}/tasks`)
    return normalizeTaskListResponse(response)
  },

  async createProjectTask(projectId: string, request: CreateProjectTaskRequest): Promise<ProjectTask> {
    const response = await requestJson<RawProjectTask>(`/api/admin/projects/${encodeURIComponent(projectId)}/tasks`, {
      method: 'POST',
      body: JSON.stringify(request),
    })
    return normalizeTask(response)
  },

  async updateProjectTask(projectId: string, taskId: string, request: UpdateProjectTaskRequest): Promise<ProjectTask> {
    const response = await requestJson<RawProjectTask>(`/api/admin/projects/${encodeURIComponent(projectId)}/tasks/${encodeURIComponent(taskId)}`, {
      method: 'PATCH',
      body: JSON.stringify(request),
    })
    return normalizeTask(response)
  },

  async deleteProjectTask(projectId: string, taskId: string): Promise<{ deleted: boolean }> {
    return requestJson<{ deleted: boolean }>(`/api/admin/projects/${encodeURIComponent(projectId)}/tasks/${encodeURIComponent(taskId)}`, {
      method: 'DELETE',
    })
  },
}
