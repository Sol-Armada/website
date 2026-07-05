<script setup lang="ts">
  import type { CreateProjectRequest, ProjectStatus, ProjectSummary } from '@/services/adminService'
  import { onMounted, ref } from 'vue'
  import { useRouter } from 'vue-router'
  import PortalShell from '@/components/layout/PortalShell.vue'
  import PageHeader from '@/components/ui/PageHeader.vue'
  import StatePanel from '@/components/ui/StatePanel.vue'
  import { adminService } from '@/services/adminService'

  const loading = ref(true)
  const error = ref<string | null>(null)
  const projects = ref<ProjectSummary[]>([])
  const showCreateModal = ref(false)
  const createLoading = ref(false)
  const createError = ref<string | null>(null)
  const projectStatuses = ref<ProjectStatus[]>([])
  const statusDropdownOpen = ref(false)
  const router = useRouter()

  const newProject = ref<CreateProjectRequest>({
    name: '',
    description: '',
    statusId: 1,
    ownerId: null,
    dueAt: null,
  })

  function getProgressColor(progress: number): string {
    if (progress >= 80) return 'text-success'
    if (progress >= 50) return 'text-primary'
    if (progress >= 25) return 'text-meta'
    return 'text-muted'
  }

  function formatDate(dateStr: string | null | undefined): string {
    if (!dateStr) return 'No due date'
    const date = new Date(dateStr)
    const now = new Date()
    const diff = date.getTime() - now.getTime()
    const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

    if (days < 0) return `${Math.abs(days)}d overdue`
    if (days === 0) return 'Due today'
    if (days === 1) return 'Due tomorrow'
    return `Due in ${days}d`
  }

  function getStatusColor(statusName?: string | null): string {
    const status = (statusName || '').toLowerCase()
    if (status.includes('done') || status.includes('complete')) return 'badge-verified'
    if (status.includes('progress') || status.includes('active')) return 'text-meta border-meta/30 bg-meta/10'
    if (status.includes('review')) return 'text-primary border-primary/30 bg-primary/10'
    if (status.includes('planning') || status.includes('backlog')) return 'badge-pending'
    return 'badge-pending'
  }

  async function loadProjects() {
    loading.value = true
    error.value = null

    try {
      const response = await adminService.listProjects()
      projects.value = response?.projects || []
    } catch(error_: any) {
      // If it's a "not found" error, just show empty state instead of error
      if (error_?.message?.toLowerCase().includes('not found') || error_?.status === 404) {
        projects.value = []
      } else {
        error.value = error_?.message || 'Failed to load projects'
      }
    } finally {
      loading.value = false
    }
  }

  async function loadProjectStatuses() {
    try {
      projectStatuses.value = await adminService.listProjectStatuses()
      if (projectStatuses.value.length > 0) {
        newProject.value.statusId = projectStatuses.value[0].id
      }
    } catch(error_: any) {
      console.error('Failed to load project statuses:', error_)
    }
  }

  function openCreateModal() {
    newProject.value = {
      name: '',
      description: '',
      statusId: projectStatuses.value.length > 0 ? projectStatuses.value[0].id : 1,
      ownerId: null,
      dueAt: null,
    }
    createError.value = null
    showCreateModal.value = true
  }

  function selectStatus(status: ProjectStatus) {
    newProject.value.statusId = status.id
    statusDropdownOpen.value = false
  }

  function getSelectedStatusName(): string {
    const selected = projectStatuses.value.find(s => s.id === newProject.value.statusId)
    return selected?.name || 'Select a status'
  }

  function closeCreateModal() {
    showCreateModal.value = false
    createError.value = null
  }

  function openProjectBoard(projectId: string) {
    void router.push(`/projects/${projectId}`)
  }

  async function handleCreateProject() {
    if (!newProject.value.name.trim()) {
      createError.value = 'Project name is required'
      return
    }

    createLoading.value = true
    createError.value = null

    try {
      const created = await adminService.createProject(newProject.value)
      projects.value.unshift(created)
      closeCreateModal()
    } catch(error_: any) {
      createError.value = error_?.message || 'Failed to create project'
    } finally {
      createLoading.value = false
    }
  }

  onMounted(async() => {
    await loadProjectStatuses()
    await loadProjects()
  })
</script>

<template>
  <PortalShell>
    <div class="w-full py-12 space-y-12">
      <div class="flex items-center justify-between">
        <PageHeader
          subtitle="Active operations, initiatives, and organizational projects"
          title="Projects"
        />

        <button
          class="btn-primary-tactical"
          type="button"
          @click="openCreateModal"
        >
          + New Project
        </button>
      </div>

      <StatePanel v-if="loading" message="Loading projects..." title="Please wait" />

      <StatePanel v-else-if="error" :message="error" title="Failed to load projects" tone="error" />

      <div v-else-if="projects.length === 0" class="tactical-panel p-12 text-center">
        <h3 class="text-xl font-semibold text-on-surface mb-2">No Projects Yet</h3>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div
          v-for="project in projects"
          :key="project.id"
          class="tactical-panel p-6 scan-hover cursor-pointer transition-all"
          @click="openProjectBoard(project.id)"
        >
          <!-- Project Header -->
          <div class="mb-4">
            <h3 class="text-xl font-semibold text-on-surface mb-2">{{ project.name }}</h3>
          </div>

          <!-- Progress Bar -->
          <div class="mb-4">
            <div class="flex justify-between text-xs mb-1">
              <span class="text-on-surface-variant">Progress</span>

              <span class="font-mono font-semibold" :class="getProgressColor(project.progress)">
                {{ project.progress }}%
              </span>
            </div>

            <div class="h-2 bg-surface-variant rounded-full overflow-hidden">
              <div
                class="h-full transition-all duration-300 rounded-full"
                :class="getProgressColor(project.progress)"
                :style="{ width: `${project.progress}%`, background: 'currentColor' }"
              />
            </div>
          </div>

          <!-- Stats Grid -->
          <div class="grid grid-cols-3 gap-4 mb-4 pb-4 border-b border-divider">

            <!-- Status Badge -->
            <div class="text-center flex items-center justify-center">
              <span class="inline-flex px-3 py-1 text-xs font-semibold rounded-md border" :class="getStatusColor(project.statusName)">
                {{ project.statusName }}
              </span>
            </div>

            <div class="text-center">
              <div class="text-xs text-on-surface-variant mb-1">Tasks</div>

              <div class="text-lg font-semibold font-mono text-on-surface">
                {{ project.doneTasks }}/{{ project.totalTasks }}
              </div>
            </div>

            <div class="text-center">
              <div class="text-xs text-on-surface-variant mb-1">Due</div>

              <div class="text-xs font-semibold font-mono text-on-surface">
                {{ formatDate(project.dueAt) }}
              </div>
            </div>
          </div>

          <!-- Footer Meta -->
          <div class="flex items-center justify-between text-xs text-on-surface-variant">
            <div v-if="project.ownerName" class="flex items-center gap-2">
              <span class="text-muted">Owner:</span>
              <span class="text-on-surface font-semibold">{{ project.ownerName }}</span>
            </div>

            <div v-else>
              <span class="text-muted">No owner</span>
            </div>
          </div>
        </div>
      </div>
      <!-- Create Project Modal -->
      <div
        v-if="showCreateModal"
        class="project-modal-overlay"
        role="presentation"
        @click.self="closeCreateModal"
      >
        <div class="project-modal-panel">
          <div class="project-modal-header">
            <div>
              <h2 class="text-lg font-semibold text-on-surface">Create New Project</h2>

              <p class="mt-1 text-sm text-on-surface-variant">
                Create a new organizational project with status tracking and task management.
              </p>
            </div>

            <button
              aria-label="Close project modal"
              class="rounded-md p-2 text-on-surface-variant transition hover:bg-surface-variant/40 hover:text-on-surface"
              type="button"
              @click="closeCreateModal"
            >
              x
            </button>
          </div>

          <form class="project-modal-body" @submit.prevent="handleCreateProject">
            <StatePanel
              v-if="createError"
              class="mb-2"
              :message="createError"
              title="Unable To Create Project"
              tone="error"
            />

            <label class="text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="project-name">
              Project Name
            </label>

            <input
              id="project-name"
              v-model="newProject.name"
              class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
              placeholder="Enter project name"
              required
              type="text"
            >

            <label class="mt-1 text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="project-description">
              Description
            </label>

            <textarea
              id="project-description"
              v-model="newProject.description"
              class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
              placeholder="Enter project description"
              rows="4"
            />

            <label class="mt-1 text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="project-status">
              Status
            </label>

            <div class="project-status-picker">
              <button
                id="project-status"
                class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface text-left w-full"
                type="button"
                @blur="statusDropdownOpen = false"
                @click="statusDropdownOpen = !statusDropdownOpen"
              >
                {{ getSelectedStatusName() }}
              </button>

              <div v-if="statusDropdownOpen" class="project-status-menu">
                <button
                  v-for="status in projectStatuses"
                  :key="status.id"
                  class="project-status-menu__item"
                  type="button"
                  @mousedown.prevent="selectStatus(status)"
                >
                  {{ status.name }}
                </button>
              </div>
            </div>

            <label class="mt-1 text-xs font-semibold uppercase tracking-wide text-on-surface-variant" for="project-due-date">
              Due Date (Optional)
            </label>

            <input
              id="project-due-date"
              v-model="newProject.dueAt"
              class="rounded-md border border-subtle bg-transparent px-3 py-2 text-sm text-on-surface"
              type="date"
            >

            <div class="mt-2 flex items-center justify-end gap-2">
              <button
                class="rounded-md border border-subtle px-4 py-2 text-sm text-on-surface transition hover:bg-surface-variant/40 disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="createLoading"
                type="button"
                @click="closeCreateModal"
              >
                Cancel
              </button>

              <button
                class="rounded-md border border-primary bg-primary px-4 py-2 text-sm font-semibold text-on-primary transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                :disabled="createLoading || !newProject.name.trim()"
                type="submit"
              >
                {{ createLoading ? 'Creating...' : 'Create Project' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </PortalShell>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.badge-verified {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(34, 197, 94, 0.1);
  border-color: rgba(34, 197, 94, 0.3);
  color: #22c55e;
}

.badge-pending {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(148, 163, 184, 0.1);
  border-color: rgba(148, 163, 184, 0.3);
  color: #94a3b8;
}

/* Project Modal Styles - Matching Attendance Modal */
.project-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgb(0 0 0 / 0.38);
  backdrop-filter: blur(4px);
}

.project-modal-panel {
  width: 100%;
  max-width: 40rem;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
  background: var(--v0-surface);
  box-shadow: 0 24px 72px rgb(0 0 0 / 0.45);
  overflow: visible;
  position: relative;
}

.project-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 1.25rem 1.25rem 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--v0-divider) 55%, transparent);
}

.project-modal-body {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem 1.25rem 1.25rem;
}

.btn-primary-tactical {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #E6A82D;
  color: #090b12;
  font-weight: 600;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s cubic-bezier(0.2, 0, 0, 1);
  min-height: 44px;
}

.btn-primary-tactical:hover {
  background: #d99920;
  box-shadow: 0 0 24px rgba(230, 168, 45, 0.3);
}

.btn-primary-tactical:active {
  transform: translateY(2px);
}

.btn-primary-tactical:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

/* Project Status Picker Dropdown */
.project-status-picker {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.project-status-menu {
  position: absolute;
  z-index: 60;
  top: calc(100% + 0.2rem);
  left: 0;
  right: 0;
  max-height: 16rem;
  overflow-y: auto;
  border: 1px solid color-mix(in srgb, var(--v0-divider) 70%, transparent);
  border-radius: 0.5rem;
  background: var(--v0-surface);
  box-shadow: 0 14px 30px rgb(0 0 0 / 0.3);
}

.project-status-menu__empty {
  padding: 0.75rem;
  text-align: center;
  font-size: 0.875rem;
  color: var(--v0-on-surface-variant);
}

.project-status-menu__item {
  width: 100%;
  padding: 0.65rem 0.875rem;
  text-align: left;
  font-size: 0.875rem;
  color: var(--v0-on-surface);
  transition: background-color 150ms ease;
  border: none;
  background: transparent;
  cursor: pointer;
}

.project-status-menu__item:hover {
  background: color-mix(in srgb, var(--v0-primary) 12%, transparent);
}

.project-status-menu__item:first-child {
  border-radius: 0.5rem 0.5rem 0 0;
}

.project-status-menu__item:last-child {
  border-radius: 0 0 0.5rem 0.5rem;
}
</style>
