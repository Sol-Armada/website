<script setup lang="ts">
  import type {
    KanbanStatus,
    MemberSummary,
    ProjectSummary,
    ProjectTask,
    ProjectTaskStatus,
    TaskPriority,
    UpdateProjectTaskRequest,
  } from '@/services/adminService'
  import StarterKit from '@tiptap/starter-kit'
  import { EditorContent, useEditor } from '@tiptap/vue-3'
  import MarkdownIt from 'markdown-it'
  import { MdEditor } from 'md-editor-v3'
  import TurndownService from 'turndown'
  import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
  import { useRoute, useRouter } from 'vue-router'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService } from '@/services/adminService'
  import 'md-editor-v3/lib/style.css'

  type KanbanTask = ProjectTask

  interface BoardColumn {
    id: KanbanStatus
    label: string
  }

  const columns = ref<BoardColumn[]>([])
  const markdownRenderer = new MarkdownIt({ breaks: true, linkify: true })
  const markdownSerializer = new TurndownService({
    headingStyle: 'atx',
    codeBlockStyle: 'fenced',
  })

  const route = useRoute()
  const router = useRouter()

  const loading = ref(true)
  const error = ref<string | null>(null)
  const project = ref<ProjectSummary | null>(null)
  const tasks = ref<KanbanTask[]>([])
  const members = ref<MemberSummary[]>([])
  const memberSearchFilter = ref('')
  const memberSearchResults = ref<MemberSummary[]>([])
  const memberSearchLoading = ref(false)
  const memberSearchDebounceTimer = ref<ReturnType<typeof setTimeout> | null>(null)
  const descriptionSaveDebounceTimer = ref<ReturnType<typeof setTimeout> | null>(null)
  const isAssigneeDropdownOpen = ref(false)
  const isStatusSelectOpen = ref(false)
  const isPrioritySelectOpen = ref(false)
  const isCreateTaskStatusDropdownOpen = ref(false)
  const isCreateTaskPriorityDropdownOpen = ref(false)
  const isNewTaskModalOpen = ref(false)
  const isTaskDetailModalOpen = ref(false)
  const selectedTaskId = ref<string | null>(null)
  const titleValidationError = ref<string | null>(null)
  const activeDescriptionEditorTaskId = ref<string | null>(null)

  const newTaskForm = ref({
    title: '',
    description: '',
    priority: 1 as TaskPriority,
    assignee: 'Unassigned',
    dueDate: '',
    status: columns.value.length > 0 ? columns.value[0].id : 'To Do' as KanbanStatus,
  })

  const draggingTaskId = ref<string | null>(null)
  const dragOverColumn = ref<KanbanStatus | null>(null)
  const isDraggingTask = ref(false)

  const routeId = (route.params as Record<string, string | string[] | undefined>).id
  const projectId = Array.isArray(routeId) ? routeId[0] : routeId
  const selectedTask = computed(() => tasks.value.find(task => task.id === selectedTaskId.value) || null)
  // const selectableParentTasks = computed(() => {
  //   if (!selectedTask.value) {
  //     return tasks.value
  //   }
  //   return tasks.value.filter(task => task.id !== selectedTask.value?.id)
  // })
  const selectedTaskChildren = computed(() => {
    if (!selectedTask.value) {
      return []
    }
    return tasks.value.filter(task => task.parentTask?.id === selectedTask.value?.id)
  })

  function isLikelyHtml(input: string): boolean {
    return /<\/?[a-z][\s\S]*>/i.test(input)
  }

  function descriptionToEditorHtml(description: string): string {
    const normalized = String(description || '').trim()
    if (!normalized) {
      return ''
    }

    if (isLikelyHtml(normalized)) {
      return normalized
    }

    return markdownRenderer.render(normalized)
  }

  function editorHtmlToMarkdown(html: string): string {
    const normalized = String(html || '').trim()
    if (!normalized || normalized === '<p></p>') {
      return ''
    }

    return markdownSerializer.turndown(normalized).trim()
  }

  const descriptionEditor = useEditor({
    extensions: [StarterKit],
    content: '',
    editorProps: {
      attributes: {
        class: 'tiptap-content',
      },
    },
    onUpdate: ({ editor }) => {
      if (!selectedTask.value) return
      const html = editor.getHTML()
      const markdown = editorHtmlToMarkdown(html)
      void updateSelectedTask('description', markdown)
    },
  })

  watch(
    () => selectedTask.value?.id ?? null,
    nextTaskId => {
      if (!descriptionEditor.value) return

      if (!nextTaskId) {
        activeDescriptionEditorTaskId.value = null
        descriptionEditor.value.commands.setContent('', { emitUpdate: false })
        return
      }

      if (activeDescriptionEditorTaskId.value === nextTaskId) {
        return
      }

      activeDescriptionEditorTaskId.value = nextTaskId
      const content = descriptionToEditorHtml(selectedTask.value?.description || '')
      descriptionEditor.value.commands.setContent(content, { emitUpdate: false })
    },
  )

  async function searchMembers(query: string) {
    // Clear existing debounce timer
    if (memberSearchDebounceTimer.value) {
      clearTimeout(memberSearchDebounceTimer.value)
    }

    // Set debounce timer for search
    memberSearchDebounceTimer.value = setTimeout(async() => {
      try {
        memberSearchLoading.value = true
        const response = await adminService.getMembers(10, 1, query || undefined)
        memberSearchResults.value = response.members || []
      } catch(error_) {
        console.error('Failed to search members:', error_)
        memberSearchResults.value = []
      } finally {
        memberSearchLoading.value = false
      }
    }, 300) // 300ms debounce
  }

  const filteredMembers = computed(() => {
    if (!memberSearchFilter.value) {
      return members.value.slice(0, 10)
    }
    return memberSearchResults.value
  })

  function taskAssigneeName(task: KanbanTask): string {
    return task.assignee?.name || task.assignee?.id || 'Unassigned'
  }

  function taskStatusName(task: KanbanTask): string {
    return task.status?.name || 'To Do'
  }

  // function taskParentTaskId(task: KanbanTask): string | null {
  //   return task.parentTask?.id || null
  // }

  function initials(name: string): string {
    return name
      .split(' ')
      .map(part => part[0])
      .join('')
      .slice(0, 2)
      .toUpperCase()
  }

  function priorityClass(priority: TaskPriority): string {
    if (priority >= 2) return 'priority-high'
    if (priority === 1) return 'priority-medium'
    return 'priority-low'
  }

  function formatDueDate(dateStr: string | null | undefined): string {
    if (!dateStr) return 'No due date'
    const date = new Date(dateStr)
    if (Number.isNaN(date.getTime())) return 'No due date'
    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
  }

  function formatDateForInput(dateStr: string | null | undefined): string {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    if (Number.isNaN(date.getTime())) return ''
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }

  function columnTasks(status: KanbanStatus): KanbanTask[] {
    return tasks.value.filter(task => taskStatusName(task) === status)
  }

  function formatStatusLabel(status: string): string {
    const normalized = status.trim()
    if (!normalized) {
      return columns.value.length > 0 ? columns.value[0].label : 'To Do'
    }

    if (normalized.toLowerCase() === 'inprogress') {
      return 'In Progress'
    }

    return normalized
      .replace(/[_-]+/g, ' ')
      .replace(/\s+/g, ' ')
      .replace(/\b\w/g, char => char.toUpperCase())
  }

  function compareTaskStatuses(a: ProjectTaskStatus, b: ProjectTaskStatus): number {
    if (a.position !== b.position) {
      return a.position - b.position
    }
    return a.name.localeCompare(b.name)
  }

  function mapStatusesToColumns(statuses: ProjectTaskStatus[]): BoardColumn[] {
    const orderedStatuses = statuses.reduce<ProjectTaskStatus[]>((acc, status) => {
      const insertAt = acc.findIndex(existing => compareTaskStatuses(status, existing) < 0)
      if (insertAt === -1) {
        acc.push(status)
      } else {
        acc.splice(insertAt, 0, status)
      }
      return acc
    }, [])

    return orderedStatuses
      .map(status => ({
        id: status.name as KanbanStatus,
        label: formatStatusLabel(status.name),
      }))
      .filter(column => column.id.trim().length > 0)
  }

  async function loadTaskStatuses() {
    if (!projectId) {
      return
    }

    try {
      const statuses = await adminService.listProjectTaskStatuses(projectId)
      const mappedColumns = mapStatusesToColumns(statuses)
      columns.value = mappedColumns.length > 0 ? mappedColumns : []
    } catch {
      columns.value = []
    }
  }

  function childTaskCount(task: KanbanTask): number {
    return tasks.value.filter(item => item.parentTask?.id === task.id).length
  }

  function toTaskUpdatePayload(task: KanbanTask): UpdateProjectTaskRequest {
    const assigneeValue = typeof task.assignee === 'string'
      ? task.assignee
      : (task.assignee?.id || '')

    const priority = task.priority == null ? 1 : Number(task.priority)

    const statusName = task.status?.name || (columns.value.length > 0 ? columns.value[0].id : 'To Do')

    return {
      title: String(task.title || '').trim(),
      description: String(task.description || '').trim(),
      priority: Number.isInteger(priority) && [0, 1, 2].includes(priority) ? priority : 1,
      assignee: String(assigneeValue).trim(),
      dueAt: task.dueAt || null,
      status: String(statusName).trim(),
      parentTaskId: task.parentTask?.id || null,
    }
  }

  async function loadTasks() {
    if (!projectId) {
      tasks.value = []
      return
    }

    const response = await adminService.listProjectTasks(projectId)
    tasks.value = response.tasks || []
  }

  async function loadMembers() {
    try {
      const response = await adminService.getMembers(100)
      members.value = response.members || []
    } catch {
      members.value = []
    }
  }

  async function persistTask(task: KanbanTask) {
    if (!projectId) {
      throw new Error('Project not found')
    }
    if (!task.id) {
      throw new Error('Task not found')
    }

    const payload = toTaskUpdatePayload(task)
    const updated = await adminService.updateProjectTask(projectId, task.id, payload)
    const index = tasks.value.findIndex(item => item.id === task.id)
    if (index !== -1) {
      tasks.value[index] = updated
    }
    if (selectedTaskId.value === updated.id) {
      selectedTaskId.value = updated.id
    }
  }

  function onDragStart(event: DragEvent, taskId: string) {
    isDraggingTask.value = true
    draggingTaskId.value = taskId

    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData('text/plain', taskId)
    }
  }

  async function onDrop(event: DragEvent, status: KanbanStatus) {
    const taskId = draggingTaskId.value || event.dataTransfer?.getData('text/plain')
    if (!taskId) return

    const task = tasks.value.find(t => t.id === taskId)
    if (!task) return

    task.status = {
      projectId: task.status?.projectId || projectId || '',
      name: status,
      position: task.status?.position || 0,
      color: task.status?.color || '',
    }
    isDraggingTask.value = false
    draggingTaskId.value = null
    dragOverColumn.value = null

    try {
      await persistTask(task)
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to move task'
      try {
        await loadTasks()
      } catch {
        // Ignore errors during reload
      }
    }
  }

  function onDragEnd() {
    isDraggingTask.value = false
    draggingTaskId.value = null
    dragOverColumn.value = null
  }

  function onTaskCardClick(taskId: string) {
    if (isDraggingTask.value) {
      return
    }
    openTaskDetail(taskId)
  }

  async function addTask() {
    if (!newTaskForm.value.title.trim()) {
      titleValidationError.value = 'Title is required'
      return
    }

    titleValidationError.value = null

    if (!projectId) {
      error.value = 'Project not found'
      return
    }

    try {
      const created = await adminService.createProjectTask(projectId, {
        title: newTaskForm.value.title.trim(),
        description: newTaskForm.value.description.trim(),
        priority: newTaskForm.value.priority,
        assignee: newTaskForm.value.assignee.trim() || 'Unassigned',
        dueAt: newTaskForm.value.dueDate || project.value?.dueAt || null,
        status: newTaskForm.value.status,
        parentTaskId: null,
      })

      tasks.value.unshift(created)
      closeNewTaskModal()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to create task'
    }
  }

  function openNewTaskModal() {
    newTaskForm.value = {
      title: '',
      description: '',
      priority: 1,
      assignee: 'Unassigned',
      dueDate: '',
      status: columns.value.length > 0 ? columns.value[0].id : 'To Do' as KanbanStatus,
    }
    titleValidationError.value = null
    isCreateTaskStatusDropdownOpen.value = false
    isCreateTaskPriorityDropdownOpen.value = false
    isNewTaskModalOpen.value = true
  }

  function closeNewTaskModal() {
    titleValidationError.value = null
    isCreateTaskStatusDropdownOpen.value = false
    isCreateTaskPriorityDropdownOpen.value = false
    isNewTaskModalOpen.value = false
  }

  function selectCreateTaskStatus(status: KanbanStatus) {
    newTaskForm.value.status = status
    isCreateTaskStatusDropdownOpen.value = false
  }

  function selectCreateTaskPriority(priority: TaskPriority) {
    newTaskForm.value.priority = priority
    isCreateTaskPriorityDropdownOpen.value = false
  }

  function closeAssigneeDropdown() {
    isAssigneeDropdownOpen.value = false
    memberSearchFilter.value = ''
  }

  // function closeSelectDropdown(selectId: string) {
  //   // Find and click the Select.Activator to toggle it closed
  //   nextTick(() => {
  //     // eslint-disable-next-line unicorn/prefer-query-selector
  //     const activator = document.getElementById(selectId)
  //     if (activator) {
  //       activator.blur()
  //       // Trigger escape key to close the select
  //       const event = new KeyboardEvent('keydown', { key: 'Escape' })
  //       activator.dispatchEvent(event)
  //     }
  //   })
  // }

  function openTaskDetail(taskId: string) {
    selectedTaskId.value = taskId
    isTaskDetailModalOpen.value = true
  }

  function closeTaskDetail() {
    selectedTaskId.value = null
    isTaskDetailModalOpen.value = false
  }

  async function updateSelectedTask(field: 'title' | 'description' | 'assignee', value: string) {
    if (!selectedTask.value) return

    if (field === 'description') {
      selectedTask.value.description = value
      const taskId = selectedTask.value.id

      if (descriptionSaveDebounceTimer.value) {
        clearTimeout(descriptionSaveDebounceTimer.value)
      }

      descriptionSaveDebounceTimer.value = setTimeout(async() => {
        descriptionSaveDebounceTimer.value = null
        const taskToPersist = tasks.value.find(task => task.id === taskId)
        if (!taskToPersist) return

        try {
          await persistTask(taskToPersist)
        } catch(error_: any) {
          error.value = error_?.message || 'Failed to update task'
          try {
            await loadTasks()
          } catch {
            // Ignore errors during reload
          }
        }
      }, 3000)

      return
    }

    if (field === 'assignee') {
      // Assignee can be empty (unassigned) or a member ID
      selectedTask.value.assignee = value.trim()
        ? { id: value.trim(), name: members.value.find(m => m.id === value.trim() || m.username === value.trim())?.username || value.trim() }
        : null
    } else {
      selectedTask.value[field] = value
    }
    try {
      await persistTask(selectedTask.value)
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to update task'
      try {
        await loadTasks()
      } catch {
        // Ignore errors during reload
      }
    }
  }

  async function updateSelectedTaskStatus(value: unknown) {
    if (!selectedTask.value) return
    selectedTask.value.status = {
      projectId: selectedTask.value.status?.projectId || projectId || '',
      name: String(value),
      position: selectedTask.value.status?.position || 0,
      color: selectedTask.value.status?.color || '',
    }
    try {
      await persistTask(selectedTask.value)
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to update task status'
      try {
        await loadTasks()
      } catch {
        // Ignore errors during reload
      }
    }
  }

  async function updateSelectedTaskPriority(value: unknown) {
    if (!selectedTask.value) return
    const parsed = Number(value)
    if (parsed >= 2) {
      selectedTask.value.priority = 2
    } else if (parsed <= 0) {
      selectedTask.value.priority = 0
    } else {
      selectedTask.value.priority = 1
    }
    try {
      await persistTask(selectedTask.value)
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to update task priority'
      try {
        await loadTasks()
      } catch {
        // Ignore errors during reload
      }
    }
  }

  async function updateSelectedTaskDueAt(value: string) {
    if (!selectedTask.value) return
    // Convert yyyy-MM-dd format to RFC3339 format
    const rfc3339Date = value
      ? new Date(`${value}T00:00:00Z`).toISOString()
      : null
    selectedTask.value.dueAt = rfc3339Date
    try {
      await persistTask(selectedTask.value)
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to update task due date'
      try {
        await loadTasks()
      } catch {
        // Ignore errors during reload
      }
    }
  }

  // async function updateSelectedTaskParent(value: unknown) {
  //   if (!selectedTask.value) return
  //   const nextParentId = String(value || '')
  //   if (nextParentId) {
  //     const parent = tasks.value.find(task => task.id === nextParentId)
  //     selectedTask.value.parentTask = parent ? { id: parent.id, title: parent.title } : { id: nextParentId }
  //   } else {
  //     selectedTask.value.parentTask = null
  //   }
  //   closeSelectDropdown('task-detail-parent')
  //   try {
  //     await persistTask(selectedTask.value)
  //   } catch(error_: any) {
  //     error.value = error_?.message || 'Failed to update parent task'
  //     try {
  //       await loadTasks()
  //     } catch {
  //       // Ignore errors during reload
  //     }
  //   }
  // }

  async function deleteSelectedTask() {
    if (!selectedTask.value) return

    if (!projectId) {
      error.value = 'Project not found'
      return
    }

    const deletingId = selectedTask.value.id

    try {
      await adminService.deleteProjectTask(projectId, deletingId)
      tasks.value = tasks.value
        .filter(task => task.id !== deletingId)
        .map(task => {
          if (task.parentTask?.id === deletingId) {
            return {
              ...task,
              parentTask: null,
            }
          }
          return task
        })
      closeTaskDetail()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to delete task'
    }
  }

  function getStatusDisplayLabel(value: unknown): string {
    const status = typeof value === 'string'
      ? value
      : (value as { name?: string } | null | undefined)?.name || ''
    return columns.value.find(column => column.id === status)?.label || formatStatusLabel(status)
  }

  function getPriorityDisplayLabel(value: unknown): string {
    const priority = Number(value ?? 1)
    if (priority >= 2) return 'High'
    if (priority <= 0) return 'Low'
    return 'Medium'
  }

  async function loadProject() {
    loading.value = true
    error.value = null

    try {
      const response = await adminService.listProjects()
      const found = response.projects.find(item => item.id === projectId)

      if (!found) {
        error.value = 'Project not found'
        return
      }

      project.value = found
      await loadTaskStatuses()
      await loadMembers()

      if (!columns.value.some(column => column.id === newTaskForm.value.status) && columns.value.length > 0) {
        newTaskForm.value.status = columns.value[0].id
      }

      await loadTasks()
    } catch(error_: any) {
      error.value = error_?.message || 'Failed to load project board'
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadProject()
  })

  onBeforeUnmount(() => {
    if (descriptionSaveDebounceTimer.value) {
      clearTimeout(descriptionSaveDebounceTimer.value)
    }
    descriptionEditor.value?.destroy()
  })
</script>

<template>
  <PortalShell>
    <div class="project-board-page w-full py-8">
      <StatePanel v-if="loading" message="Preparing project Kanban..." title="Loading board" />

      <StatePanel v-else-if="error" :message="error" title="Board unavailable" tone="error" />

      <template v-else-if="project">
        <div class="project-board-layout">
          <div class="project-header">
            <button class="back-link" type="button" @click="router.push('/projects')">
              ← Back to Projects
            </button>

            <h1 class="project-title">{{ project.name }}</h1>

            <div class="project-meta">
              <span>{{ project.statusName }}</span>
              <span>•</span>
              <span>Due: {{ formatDueDate(project.dueAt) }}</span>
            </div>
          </div>

          <div class="project-actions">
            <button class="btn-primary" type="button" @click="openNewTaskModal">
              + New Task
            </button>
          </div>

          <div class="kanban-board">
            <div
              v-for="column in columns"
              :key="column.id"
              class="kanban-column"
              @dragenter.prevent="dragOverColumn = column.id"
              @dragover.prevent
              @drop.prevent="onDrop($event, column.id)"
            >
              <div class="column-header">
                <span class="column-title">{{ column.label }}</span>
                <span class="column-count">{{ columnTasks(column.id).length }}</span>
              </div>

              <div
                class="column-cards"
                :class="{ 'drag-over': dragOverColumn === column.id }"
                @dragenter.prevent="dragOverColumn = column.id"
                @dragleave="dragOverColumn = null"
                @dragover.prevent
                @drop.prevent="onDrop($event, column.id)"
              >
                <div
                  v-for="task in columnTasks(column.id)"
                  :key="task.id"
                  class="task-card"
                  draggable="true"
                  @click="onTaskCardClick(task.id)"
                  @dragend="onDragEnd"
                  @dragenter.prevent="dragOverColumn = column.id"
                  @dragover.prevent
                  @dragstart="onDragStart($event, task.id)"
                  @drop.prevent.stop="onDrop($event, column.id)"
                >
                  <div class="task-header">
                    <div>
                      <div class="task-title">{{ task.title }}</div>

                      <div class="task-id">#{{ task.id }}</div>
                    </div>

                    <span class="priority-badge" :class="priorityClass(task.priority)">
                      {{ getPriorityDisplayLabel(task.priority) }}
                    </span>
                  </div>

                  <div class="task-meta">
                    <div class="assignee">
                      <span class="avatar">{{ initials(taskAssigneeName(task)) }}</span>
                      <span class="due-date">{{ formatDueDate(task.dueAt) }}</span>
                    </div>

                    <span class="task-count">{{ childTaskCount(task) }} child tasks</span>
                  </div>
                </div>

                <div v-if="columnTasks(column.id).length === 0" class="empty-state">
                  <div class="empty-state-text">No tasks</div>
                </div>

                <div
                  v-show="isDraggingTask"
                  class="column-drop-zone"
                  @dragenter.prevent="dragOverColumn = column.id"
                  @dragover.prevent
                  @drop.prevent.stop="onDrop($event, column.id)"
                >
                  <span class="column-drop-zone-label">Drop task here</span>
                </div>
              </div>
            </div>
          </div>

          <div
            v-if="isTaskDetailModalOpen && selectedTask"
            class="task-detail-overlay"
            role="presentation"
            @click.self="closeTaskDetail"
          >
            <div class="task-detail-panel">
              <div class="task-detail-body">
                <div class="task-detail-main">
                  <input
                    class="task-detail-title"
                    type="text"
                    :value="selectedTask.title"
                    @input="updateSelectedTask('title', ($event.target as HTMLInputElement).value)"
                  >

                  <div class="task-detail-description tiptap-editor-shell">
                    <EditorContent v-if="descriptionEditor" :editor="descriptionEditor" />
                  </div>

                  <section class="task-section">
                    <h4 class="task-section-title">Child Tasks</h4>

                    <div class="task-child-list">
                      <article v-for="child in selectedTaskChildren" :key="child.id" class="task-child-item">
                        <div>
                          <div class="task-child-title">{{ child.title }}</div>
                          <div class="task-child-meta">#{{ child.id }} • {{ getStatusDisplayLabel(child.status) }}</div>
                        </div>

                        <button class="task-child-open" type="button" @click="openTaskDetail(child.id)">
                          Open
                        </button>
                      </article>

                      <p v-if="selectedTaskChildren.length === 0" class="task-child-empty">
                        No child tasks linked.
                      </p>
                    </div>
                  </section>

                  <section class="task-section">
                    <h4 class="task-section-title">Activity</h4>

                    <div class="task-activity-list">
                      <article
                        v-for="(entry, index) in (selectedTask.activity || [])"
                        :key="entry.id || String(index)"
                        class="task-activity-item"
                      >
                        <p class="task-activity-details">{{ entry.summary }}</p>
                      </article>
                    </div>
                  </section>
                </div>

                <aside class="task-detail-sidebar">

                  <div class="task-sidebar-field">
                    <label class="task-sidebar-label" for="task-detail-status">Status</label>

                    <div class="assignee-combobox">
                      <button
                        id="task-detail-status"
                        class="assignee-input"
                        type="button"
                        @blur="isStatusSelectOpen = false"
                        @focus="isStatusSelectOpen = true"
                      >
                        {{ getStatusDisplayLabel(selectedTask.status?.name || (columns.length > 0 ? columns[0].id : '')) }}
                      </button>

                      <div v-if="isStatusSelectOpen" class="assignee-dropdown">
                        <div
                          v-for="column in columns"
                          :id="`detail-status-${column.id}`"
                          :key="column.id"
                          class="assignee-option"
                          @mousedown.prevent="updateSelectedTaskStatus(column.id); isStatusSelectOpen = false"
                        >
                          {{ column.label }}
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="task-sidebar-field">
                    <label class="task-sidebar-label" for="task-detail-priority">Priority</label>

                    <div class="assignee-combobox">
                      <button
                        id="task-detail-priority"
                        class="assignee-input"
                        type="button"
                        @blur="isPrioritySelectOpen = false"
                        @focus="isPrioritySelectOpen = true"
                      >
                        {{ getPriorityDisplayLabel(selectedTask.priority) }}
                      </button>

                      <div v-if="isPrioritySelectOpen" class="assignee-dropdown">
                        <div
                          id="detail-priority-low"
                          class="assignee-option"
                          @mousedown.prevent="updateSelectedTaskPriority(0); isPrioritySelectOpen = false"
                        >
                          Low
                        </div>

                        <div
                          id="detail-priority-medium"
                          class="assignee-option"
                          @mousedown.prevent="updateSelectedTaskPriority(1); isPrioritySelectOpen = false"
                        >
                          Medium
                        </div>

                        <div
                          id="detail-priority-high"
                          class="assignee-option"
                          @mousedown.prevent="updateSelectedTaskPriority(2); isPrioritySelectOpen = false"
                        >
                          High
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="task-sidebar-field">
                    <label class="task-sidebar-label" for="task-detail-assignee">Assigned To</label>

                    <div class="assignee-combobox">
                      <input
                        v-if="selectedTask"
                        id="task-detail-assignee"
                        class="assignee-input"
                        :placeholder="selectedTask.assignee ? (members.find(m => m.id === selectedTask?.assignee?.id || m.username === selectedTask?.assignee?.name)?.username || 'Unassigned') : 'Unassigned'"
                        type="text"
                        :value="memberSearchFilter"
                        @blur="closeAssigneeDropdown()"
                        @focus="isAssigneeDropdownOpen = true; if (!memberSearchFilter) { memberSearchResults = members.slice(0, 10) }"
                        @input="memberSearchFilter = ($event.target as HTMLInputElement).value; searchMembers(memberSearchFilter)"
                      >

                      <div v-if="isAssigneeDropdownOpen" class="assignee-dropdown">
                        <div v-if="memberSearchLoading" class="assignee-loading">
                          Searching...
                        </div>

                        <template v-else>
                          <div
                            id="detail-assignee-none"
                            class="assignee-option"
                            @mousedown.prevent="updateSelectedTask('assignee', ''); closeAssigneeDropdown()"
                          >
                            Unassigned
                          </div>

                          <div
                            v-for="member in filteredMembers"
                            :id="`detail-assignee-${member.id}`"
                            :key="member.id"
                            class="assignee-option"
                            @mousedown.prevent="updateSelectedTask('assignee', member.id); closeAssigneeDropdown()"
                          >
                            {{ member.username }}
                          </div>

                          <div v-if="filteredMembers.length === 0 && memberSearchFilter" class="assignee-empty">
                            No members found
                          </div>
                        </template>
                      </div>
                    </div>
                  </div>

                  <div class="task-sidebar-field">
                    <label class="task-sidebar-label" for="task-detail-due">Due Date</label>

                    <input
                      id="task-detail-due"
                      class="task-field"
                      type="date"
                      :value="formatDateForInput(selectedTask.dueAt)"
                      @input="updateSelectedTaskDueAt(($event.target as HTMLInputElement).value)"
                    >
                  </div>

                  <div class="task-sidebar-actions">
                    <button class="task-delete-btn" type="button" @click="deleteSelectedTask">
                      Delete Task
                    </button>
                  </div>
                </aside>
              </div>

              <div class="task-detail-footer">
                <span class="task-detail-footer-id">Task #{{ selectedTask.id }}</span>

                <button class="task-detail-footer-close" type="button" @click="closeTaskDetail">
                  Close
                </button>
              </div>
            </div>
          </div>

          <div
            v-if="isNewTaskModalOpen"
            class="task-modal-overlay"
            role="presentation"
            @click.self="closeNewTaskModal"
          >
            <div class="task-modal-panel">
              <div class="task-modal-header">
                <h3 class="task-modal-title">Create Task</h3>

                <button
                  aria-label="Close new task modal"
                  class="task-modal-close"
                  type="button"
                  @click="closeNewTaskModal"
                >
                  x
                </button>
              </div>

              <form class="task-modal-body" @submit.prevent="addTask">
                <label class="task-field-label" for="task-title">Title *</label>

                <input
                  id="task-title"
                  v-model="newTaskForm.title"
                  class="task-field"
                  placeholder="Enter task title"
                  required
                  type="text"
                >

                <p v-if="titleValidationError" class="task-field-error">{{ titleValidationError }}</p>

                <label class="task-field-label" for="task-description">Description</label>

                <MdEditor
                  id="task-description"
                  v-model="newTaskForm.description"
                  class="task-field task-field-textarea"
                  :preview="false"
                />

                <div class="task-field-row">
                  <div>
                    <label class="task-field-label" for="task-status">Status</label>

                    <div class="assignee-combobox">
                      <button
                        id="task-status"
                        class="task-select-trigger"
                        type="button"
                        @blur="isCreateTaskStatusDropdownOpen = false"
                        @focus="isCreateTaskStatusDropdownOpen = true"
                      >
                        <span>{{ getStatusDisplayLabel(newTaskForm.status) }}</span>
                        <span class="task-select-chevron">⌄</span>
                      </button>

                      <div v-if="isCreateTaskStatusDropdownOpen" class="task-select-menu">
                        <div
                          v-for="column in columns"
                          :id="`status-${column.id}`"
                          :key="column.id"
                          class="task-select-option"
                          @mousedown.prevent="selectCreateTaskStatus(column.id)"
                        >
                          {{ column.label }}
                        </div>
                      </div>
                    </div>
                  </div>

                  <div>
                    <label class="task-field-label" for="task-priority">Priority</label>

                    <div class="assignee-combobox">
                      <button
                        id="task-priority"
                        class="task-select-trigger"
                        type="button"
                        @blur="isCreateTaskPriorityDropdownOpen = false"
                        @focus="isCreateTaskPriorityDropdownOpen = true"
                      >
                        <span>{{ getPriorityDisplayLabel(newTaskForm.priority) }}</span>
                        <span class="task-select-chevron">⌄</span>
                      </button>

                      <div v-if="isCreateTaskPriorityDropdownOpen" class="task-select-menu">
                        <div id="priority-low" class="task-select-option" @mousedown.prevent="selectCreateTaskPriority(0)">
                          Low
                        </div>

                        <div id="priority-medium" class="task-select-option" @mousedown.prevent="selectCreateTaskPriority(1)">
                          Medium
                        </div>

                        <div id="priority-high" class="task-select-option" @mousedown.prevent="selectCreateTaskPriority(2)">
                          High
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <label class="task-field-label" for="task-assignee">Assignee</label>

                <input
                  id="task-assignee"
                  v-model="newTaskForm.assignee"
                  class="task-field"
                  placeholder="Assignee name"
                  type="text"
                >

                <label class="task-field-label" for="task-due-date">Due Date (optional)</label>

                <input
                  id="task-due-date"
                  v-model="newTaskForm.dueDate"
                  class="task-field"
                  type="date"
                >

                <div class="task-modal-actions">
                  <button class="btn-secondary" type="button" @click="closeNewTaskModal">
                    Cancel
                  </button>

                  <button class="btn-primary" :disabled="!newTaskForm.title.trim()" type="submit">
                    Save Task
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </template>
    </div>
  </PortalShell>
</template>

<style scoped>
  .project-board-page {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    min-height: calc(100dvh - 13rem);
  }

  .project-board-layout {
    display: flex;
    flex: 1;
    min-height: 0;
    flex-direction: column;
    gap: 1rem;
  }

  .project-header {
    margin-bottom: 0;
  }

  .project-actions {
    flex: 0 0 auto;
  }

  .back-link {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--sa-gold);
    text-decoration: none;
    font-size: 0.875rem;
    margin-bottom: 1rem;
    transition: gap 150ms cubic-bezier(0.2, 0, 0, 1);
    min-height: 44px;
    border-radius: 10px;
    padding: 0.5rem;
    margin-left: -0.5rem;
  }

  .back-link:hover {
    gap: 0.75rem;
  }

  .project-title {
    font-size: 2.25rem;
    color: var(--sa-gold);
    margin-bottom: 0.75rem;
    font-weight: 700;
  }

  .project-meta {
    display: flex;
    gap: 1rem;
    color: var(--sa-muted);
    font-size: 0.875rem;
    flex-wrap: wrap;
  }

  .btn-primary {
    background: var(--sa-gold);
    color: var(--sa-bg);
    border: none;
    padding: 0.75rem 1.25rem;
    border-radius: 10px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 150ms cubic-bezier(0.2, 0, 0, 1);
    min-height: 44px;
  }

  .btn-primary:hover {
    box-shadow: 0 0 24px rgba(230, 168, 45, 0.4);
  }

  .btn-secondary {
    background: transparent;
    color: var(--sa-fg);
    border: 1px solid var(--sa-border);
    padding: 0.75rem 1.25rem;
    border-radius: 10px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 150ms cubic-bezier(0.2, 0, 0, 1);
    min-height: 44px;
  }

  .btn-secondary:hover {
    border-color: var(--sa-gold);
    color: var(--sa-gold);
  }

  .kanban-board {
    display: flex;
    flex: 1;
    min-height: 0;
    gap: 1.5rem;
    overflow-x: auto;
    padding: 0.25rem 0 1rem;
  }

  .kanban-column {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-width: 300px;
    flex-shrink: 0;
  }

  .column-header {
    background: var(--sa-surface);
    border: 1px solid var(--sa-border);
    border-radius: 16px;
    padding: 1rem;
    margin-bottom: 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    position: relative;
    overflow: hidden;
  }

  .column-header::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: linear-gradient(90deg, var(--sa-gold) 0%, var(--sa-blue) 50%, transparent 100%);
    opacity: 0.8;
  }

  .column-title {
    font-size: 0.875rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-family: var(--font-mono);
  }

  .column-count {
    background: rgba(230, 168, 45, 0.1);
    color: var(--sa-gold);
    padding: 0.1rem 0.5rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }

  .column-cards {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.75rem;
    min-height: 0;
    padding: 0.25rem;
    border-radius: 16px;
    overflow-y: auto;
    transition: background 150ms cubic-bezier(0.2, 0, 0, 1);
  }

  .column-cards.drag-over {
    background: rgba(230, 168, 45, 0.08);
    border: 1px dashed var(--sa-gold);
  }

  .column-drop-zone {
    position: sticky;
    bottom: 0;
    z-index: 1;
    flex: 0 0 auto;
    min-height: 5.5rem;
    border: 1px dashed rgb(230 168 45 / 0.25);
    border-radius: 0.5rem;
    background: rgb(230 168 45 / 0.05);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 0.25rem;
    backdrop-filter: blur(2px);
  }

  .column-drop-zone-label {
    color: var(--sa-gold);
    font-size: 0.75rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    font-family: var(--font-mono);
    opacity: 0.82;
  }

  .column-cards.drag-over .column-drop-zone {
    border-color: rgb(230 168 45 / 0.7);
    background: rgb(230 168 45 / 0.16);
  }

  .column-cards.drag-over .column-drop-zone-label {
    opacity: 1;
  }

  .task-card {
    background: var(--sa-surface);
    border: 1px solid var(--sa-border);
    border-radius: 10px;
    padding: 0.9rem;
    cursor: move;
    transition: all 150ms cubic-bezier(0.2, 0, 0, 1);
  }

  .task-card:hover {
    border-color: var(--sa-gold);
    box-shadow: 0 0 16px rgba(230, 168, 45, 0.2);
    transform: translateY(-2px);
  }

  .task-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .task-title {
    font-size: 0.94rem;
    font-weight: 600;
    color: var(--sa-fg);
  }

  .task-description {
    margin-top: 0.25rem;
    color: var(--sa-fg-2);
    font-size: 0.78rem;
    line-height: 1.35;
  }

  .markdown-card-preview {
    max-height: 5.25rem;
    overflow: hidden;
  }

  :deep(.markdown-card-preview .md-editor-preview-wrapper) {
    padding: 0;
    background: transparent;
  }

  :deep(.markdown-card-preview .md-editor-preview) {
    color: var(--sa-fg-2);
    font-size: 0.78rem;
    line-height: 1.35;
  }

  :deep(.markdown-card-preview .md-editor-preview h1),
  :deep(.markdown-card-preview .md-editor-preview h2),
  :deep(.markdown-card-preview .md-editor-preview h3) {
    margin: 0.1rem 0 0.35rem;
    color: var(--sa-fg);
    font-size: 0.9rem;
  }

  .markdown-card-preview {
    max-height: 5.25rem;
    overflow: hidden;
  }

  :deep(.markdown-card-preview .md-editor-preview-wrapper) {
    padding: 0;
    background: transparent;
  }

  :deep(.markdown-card-preview p) {
    margin: 0 0 0.3rem;
  }

  .task-id {
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--sa-muted);
  }

  .priority-badge {
    padding: 0.1rem 0.45rem;
    border-radius: 6px;
    font-size: 0.65rem;
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .priority-high {
    background: rgba(251, 113, 133, 0.1);
    border: 1px solid rgba(251, 113, 133, 0.3);
    color: var(--sa-danger);
  }

  .priority-medium {
    background: rgba(96, 165, 250, 0.1);
    border: 1px solid rgba(96, 165, 250, 0.3);
    color: #60a5fa;
  }

  .priority-low {
    background: rgba(148, 163, 184, 0.1);
    border: 1px solid rgba(148, 163, 184, 0.3);
    color: var(--sa-muted);
  }

  .task-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.65rem;
    padding-top: 0.65rem;
    border-top: 1px solid var(--sa-border-soft);
  }

  .assignee {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .avatar {
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--sa-gold), var(--sa-blue));
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 0.58rem;
    font-weight: 700;
    color: var(--sa-bg);
  }

  .due-date,
  .task-count {
    font-family: var(--font-mono);
    font-size: 0.68rem;
    color: var(--sa-muted);
  }

  .empty-state {
    padding: 2rem 1rem;
    text-align: center;
    color: var(--sa-muted);
  }

  .empty-state-text {
    font-size: 0.75rem;
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .task-modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 70;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: rgb(0 0 0 / 0.5);
    backdrop-filter: blur(4px);
  }

  .task-modal-panel {
    width: 100%;
    max-width: 40rem;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    background: var(--v0-surface);
    box-shadow: 0 24px 72px rgb(0 0 0 / 0.45);
    overflow: hidden;
  }

  .task-modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.25rem;
    border-bottom: 1px solid color-mix(in srgb, var(--v0-divider) 55%, transparent);
  }

  .task-modal-title {
    color: var(--sa-gold);
    font-size: 1.125rem;
    font-weight: 700;
  }

  .task-modal-close {
    background: none;
    border: none;
    color: var(--sa-muted);
    width: 2rem;
    height: 2rem;
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .task-modal-close:hover {
    background: rgb(148 163 184 / 0.14);
    color: var(--sa-fg);
  }

  .task-modal-body {
    display: flex;
    flex-direction: column;
    gap: 0.65rem;
    padding: 1.25rem;
  }

  .task-field-label {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--sa-muted);
    font-family: var(--font-mono);
  }

  .task-field-error {
    margin-top: -0.25rem;
    font-size: 0.75rem;
    color: var(--sa-danger);
  }

  .task-field-row {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.75rem;
  }

  .task-field {
    width: 100%;
    border-radius: 0.625rem;
    border: 1px solid var(--sa-border);
    background: transparent;
    color: var(--sa-fg);
    padding: 0.6rem 0.8rem;
    font-size: 0.875rem;
  }

  .task-field:focus {
    outline: none;
    border-color: var(--sa-gold);
    box-shadow: 0 0 0 2px rgb(230 168 45 / 0.2);
  }

  .task-field-textarea {
    min-height: 13rem;
  }

  :deep(.task-field-textarea .md-editor-toolbar-wrapper),
  :deep(.task-detail-description .md-editor-toolbar-wrapper) {
    display: none;
  }

  :deep(.task-field-textarea .md-editor-footer),
  :deep(.task-detail-description .md-editor-footer) {
    display: none;
  }

  :deep(.task-field-textarea.md-editor),
  :deep(.task-detail-description.md-editor) {
    background: color-mix(in srgb, var(--v0-surface) 90%, rgb(0 0 0 / 1));
    border: none;
  }

  :deep(.task-field-textarea .md-editor-input-wrapper),
  :deep(.task-detail-description .md-editor-input-wrapper) {
    background: color-mix(in srgb, var(--v0-surface) 90%, rgb(0 0 0 / 1));
  }

  :deep(.task-detail-description .md-editor-preview-wrapper),
  :deep(.task-field-textarea .md-editor-preview-wrapper) {
    background: color-mix(in srgb, var(--v0-surface) 90%, rgb(0 0 0 / 1));
    border-left: 1px solid var(--sa-border-soft);
  }

  :deep(.task-detail-description .md-editor-preview),
  :deep(.task-field-textarea .md-editor-preview) {
    color: var(--sa-fg);
  }

  :deep(.task-field-textarea .cm-editor),
  :deep(.task-detail-description .cm-editor),
  :deep(.task-field-textarea .cm-scroller),
  :deep(.task-detail-description .cm-scroller),
  :deep(.task-field-textarea .cm-content),
  :deep(.task-detail-description .cm-content) {
    background: color-mix(in srgb, var(--v0-surface) 90%, rgb(0 0 0 / 1));
  }

  :deep(.task-field-textarea .md-editor-input),
  :deep(.task-detail-description .md-editor-input) {
    color: var(--sa-fg);
  }

  :deep(.task-field-textarea .cm-content),
  :deep(.task-detail-description .cm-content),
  :deep(.task-field-textarea .cm-line),
  :deep(.task-detail-description .cm-line),
  :deep(.task-field-textarea .cm-gutters),
  :deep(.task-detail-description .cm-gutters) {
    color: var(--sa-fg);
  }

  .task-select-trigger {
    width: 100%;
    border-radius: 0.625rem;
    border: 1px solid var(--sa-border);
    background: transparent;
    color: var(--sa-fg);
    padding: 0.6rem 0.8rem;
    font-size: 0.875rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    cursor: pointer;
  }

  .task-select-trigger:hover {
    border-color: var(--sa-gold);
  }

  .task-select-trigger:focus {
    outline: none;
    border-color: var(--sa-gold);
    box-shadow: 0 0 0 2px rgb(230 168 45 / 0.2);
  }

  .task-select-chevron {
    color: var(--sa-muted);
    font-size: 0.95rem;
    line-height: 1;
  }

  .task-select-menu {
    position: absolute;
    top: calc(100% + 0.25rem);
    left: 0;
    right: 0;
    z-index: 80;
    max-height: 14rem;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    border-radius: 0.5rem;
    background: var(--v0-surface);
    box-shadow: 0 14px 30px rgb(0 0 0 / 0.3);
    display: flex;
    flex-direction: column;
  }

  .task-select-option {
    width: 100%;
    border: none;
    background: transparent;
    color: var(--sa-fg);
    padding: 0.65rem 0.85rem;
    text-align: left;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background-color 150ms ease;
  }

  .task-select-option:hover {
    background: color-mix(in srgb, var(--v0-primary) 12%, transparent);
  }

  .task-select-search {
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: none;
    border-bottom: 1px solid color-mix(in srgb, var(--v0-divider) 50%, transparent);
    background: transparent;
    color: var(--sa-fg-2);
    font-size: 0.875rem;
    outline: none;
    flex-shrink: 0;
  }

  .task-select-search:focus {
    border-color: var(--sa-gold);
    background: rgb(230 168 45 / 0.05);
  }

  .task-select-search::placeholder {
    color: var(--sa-muted);
  }

  .task-select-menu > :not(.task-select-search) {
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }

  .assignee-combobox {
    position: relative;
  }

  .assignee-input {
    width: 100%;
    border: 1px solid var(--sa-border);
    border-radius: 0.625rem;
    background: transparent;
    color: var(--sa-fg);
    padding: 0.6rem 0.8rem;
    font-size: 0.875rem;
    outline: none;
  }

  .assignee-input:hover {
    border-color: var(--sa-gold);
  }

  .assignee-input:focus {
    outline: none;
    border-color: var(--sa-gold);
    box-shadow: 0 0 0 2px rgb(230 168 45 / 0.2);
  }

  .assignee-input::placeholder {
    color: var(--sa-muted);
  }

  .assignee-input[type="button"] {
    text-align: left;
    padding: 0.6rem 0.8rem;
    font-family: inherit;
    cursor: pointer;
    user-select: none;
  }

  .assignee-input[type="button"]:hover {
    border-color: var(--sa-gold);
  }

  .assignee-input[type="button"]:focus {
    outline: none;
    border-color: var(--sa-gold);
    box-shadow: 0 0 0 2px rgb(230 168 45 / 0.2);
  }

  .assignee-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    z-index: 80;
    margin-top: 0.25rem;
    max-height: 14rem;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    border-radius: 0.5rem;
    background: var(--v0-surface);
    box-shadow: 0 14px 30px rgb(0 0 0 / 0.3);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }

  .assignee-option {
    width: 100%;
    border: none;
    background: transparent;
    color: var(--sa-fg);
    padding: 0.65rem 0.85rem;
    text-align: left;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background-color 150ms ease;
    flex-shrink: 0;
  }

  .assignee-option:hover {
    background: color-mix(in srgb, var(--v0-primary) 12%, transparent);
  }

  .assignee-loading {
    padding: 1rem 0.85rem;
    text-align: center;
    color: var(--sa-muted);
    font-size: 0.875rem;
  }

  .assignee-empty {
    padding: 1rem 0.85rem;
    text-align: center;
    color: var(--sa-muted);
    font-size: 0.875rem;
  }

  .task-modal-actions {
    margin-top: 0.75rem;
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .task-detail-overlay {
    position: fixed;
    inset: 0;
    z-index: 75;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
    background: rgb(0 0 0 / 0.6);
    backdrop-filter: blur(4px);
  }

  .task-detail-panel {
    width: 100%;
    max-width: 1100px;
    max-height: 92vh;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
    background: var(--v0-surface);
    box-shadow: 0 24px 72px rgb(0 0 0 / 0.45);
    /* overflow: hidden; */
    display: flex;
    flex-direction: column;
  }

  .task-detail-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.25rem;
    border-bottom: 1px solid color-mix(in srgb, var(--v0-divider) 55%, transparent);
  }

  .task-detail-header h3 {
    color: var(--sa-gold);
    font-size: 2rem;
    font-weight: 700;
    font-family: var(--font-mono);
  }

  .task-detail-close {
    background: none;
    border: none;
    color: var(--sa-muted);
    width: 2rem;
    height: 2rem;
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .task-detail-close:hover {
    background: rgb(148 163 184 / 0.14);
    color: var(--sa-fg);
  }

  .task-detail-body {
    display: flex;
    min-height: 0;
    flex: 1;
  }

  .task-detail-main {
    flex: 1;
    min-width: 0;
    padding: 1.25rem;
    overflow-y: auto;
  }

  .task-detail-title {
    width: 100%;
    margin-bottom: 0.9rem;
    font-size: 2rem;
    font-weight: 700;
    color: var(--sa-fg);
    border: none;
    background: transparent;
    border-radius: 0.5rem;
    padding: 0.4rem 0.5rem;
  }

  .task-detail-title:hover,
  .task-detail-title:focus {
    outline: none;
    background: rgb(230 168 45 / 0.08);
  }

  .task-detail-description {
    width: 100%;
    margin-bottom: 1.1rem;
    min-height: 17rem;
    border: 1px solid var(--sa-border);
    border-radius: 0.625rem;
    overflow: hidden;
  }

  .tiptap-editor-shell {
    background: color-mix(in srgb, var(--v0-surface) 90%, rgb(0 0 0 / 1));
    padding: 0.85rem 0.95rem;
  }

  :deep(.tiptap-editor-shell .tiptap) {
    min-height: 15rem;
    color: var(--sa-fg);
    outline: none;
    white-space: pre-wrap;
  }

  :deep(.tiptap-editor-shell .tiptap h1) {
    font-size: 1.45rem;
    line-height: 1.25;
    margin: 0.45rem 0;
  }

  :deep(.tiptap-editor-shell .tiptap h2) {
    font-size: 1.2rem;
    line-height: 1.3;
    margin: 0.4rem 0;
  }

  :deep(.tiptap-editor-shell .tiptap h3) {
    font-size: 1.05rem;
    line-height: 1.35;
    margin: 0.35rem 0;
  }

  :deep(.tiptap-editor-shell .tiptap p) {
    margin: 0.35rem 0;
  }

  :deep(.tiptap-editor-shell .tiptap ul),
  :deep(.tiptap-editor-shell .tiptap ol) {
    margin: 0.4rem 0;
    padding-left: 1.2rem;
  }

  .task-detail-description:focus {
    outline: none;
    border-color: var(--sa-gold);
    box-shadow: 0 0 0 2px rgb(230 168 45 / 0.2);
  }

  .task-detail-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.9rem 1.25rem;
    border-top: 1px solid var(--sa-border-soft);
    background: rgb(0 0 0 / 0.14);
  }

  .task-detail-footer-id {
    color: var(--sa-muted);
    font-size: 0.78rem;
    font-family: var(--font-mono);
    letter-spacing: 0.03em;
  }

  .task-detail-footer-close {
    border: 1px solid var(--sa-border);
    background: transparent;
    color: var(--sa-fg);
    border-radius: 0.5rem;
    padding: 0.45rem 0.85rem;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 150ms cubic-bezier(0.2, 0, 0, 1);
  }

  .task-detail-footer-close:hover {
    border-color: var(--sa-gold);
    color: var(--sa-gold);
  }

  .task-section {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--sa-border-soft);
  }

  .task-section-title {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--sa-muted);
    font-family: var(--font-mono);
    margin-bottom: 0.65rem;
  }

  .task-child-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .task-child-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    background: rgb(0 0 0 / 0.16);
    border: 1px solid var(--sa-border);
    border-radius: 0.5rem;
    padding: 0.55rem 0.65rem;
  }

  .task-child-title {
    color: var(--sa-fg);
    font-size: 0.88rem;
    font-weight: 600;
  }

  .task-child-meta {
    color: var(--sa-muted);
    font-size: 0.74rem;
    font-family: var(--font-mono);
    margin-top: 0.1rem;
  }

  .task-child-open {
    border: 1px solid var(--sa-border);
    background: transparent;
    color: var(--sa-fg);
    border-radius: 0.5rem;
    padding: 0.35rem 0.65rem;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
  }

  .task-child-open:hover {
    border-color: var(--sa-gold);
    color: var(--sa-gold);
  }

  .task-child-empty {
    color: var(--sa-muted);
    font-size: 0.82rem;
    font-style: italic;
    padding: 0.2rem 0;
  }

  .task-activity-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .task-activity-item {
    display: flex;
    gap: 0.5rem;
  }

  .task-activity-avatar {
    width: 1.8rem;
    height: 1.8rem;
    border-radius: 999px;
    background: var(--sa-gold);
    color: #06101d;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 0.68rem;
    font-weight: 700;
    font-family: var(--font-mono);
    flex-shrink: 0;
  }

  .task-activity-content {
    flex: 1;
    min-width: 0;
  }

  .task-activity-header {
    display: flex;
    align-items: baseline;
    gap: 0.35rem;
    flex-wrap: wrap;
  }

  .task-activity-name {
    color: var(--sa-fg);
    font-size: 0.86rem;
    font-weight: 600;
  }

  .task-activity-action,
  .task-activity-time {
    color: var(--sa-muted);
    font-size: 0.8rem;
  }

  .task-activity-time {
    margin-left: auto;
  }

  .task-activity-details {
    /* margin-top: 0.35rem; */
    color: var(--sa-fg-2);
    font-size: 0.69rem;
    /* background: rgb(0 0 0 / 0.16); */
    /* border: 1px solid var(--sa-border); */
    /* border-radius: 0.5rem; */
    /* padding: 0.45rem 0.6rem; */
  }

  .task-detail-sidebar {
    width: 260px;
    border-left: 1px solid var(--sa-border-soft);
    background: rgb(0 0 0 / 0.18);
    padding: 1rem;
    /* overflow-y: auto; */
  }

  .task-sidebar-field {
    margin-bottom: 0.95rem;
  }

  .task-sidebar-label {
    display: block;
    margin-bottom: 0.35rem;
    font-size: 0.7rem;
    color: var(--sa-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-family: var(--font-mono);
    font-weight: 700;
  }

  .task-sidebar-actions {
    border-top: 1px solid var(--sa-border-soft);
    padding-top: 0.95rem;
    margin-top: 1rem;
  }

  .task-delete-btn {
    width: 100%;
    border: 1px solid var(--sa-danger);
    color: var(--sa-danger);
    background: transparent;
    border-radius: 0.625rem;
    padding: 0.65rem 0.8rem;
    font-size: 0.82rem;
    font-weight: 700;
    cursor: pointer;
  }

  .task-delete-btn:hover {
    background: rgb(251 113 133 / 0.12);
  }

  @media (max-width: 920px) {
    .project-board-page {
      min-height: auto;
    }

    .project-board-layout {
      min-height: auto;
    }

    .project-title {
      font-size: 1.5rem;
    }

    .kanban-column {
      min-width: 280px;
    }

    .task-field-row {
      grid-template-columns: 1fr;
    }

    .task-detail-panel {
      max-height: 100vh;
      border-radius: 0;
    }

    .task-detail-body {
      flex-direction: column;
    }

    .task-detail-sidebar {
      width: 100%;
      border-left: none;
      border-top: 1px solid var(--sa-border-soft);
    }
  }
</style>
