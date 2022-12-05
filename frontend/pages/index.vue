<script setup lang="ts">
import { Task } from '@/graphql/generated'
const { $graphql } = useAPI()

const prefix = 'Index'

const { getMe, logOut } = useSession()
const me = await getMe()

const tasks = useState<Task[]>(`${prefix}.tasks`, () => [])

const refreshTasks = () => $graphql
  .tasksByCreator({ creatorUserId: me.value.id })
  .then(resp => tasks.value = resp.tasksByCreator)

await refreshTasks()

const addTask = () => {
  $graphql.createTask({})
    .then(resp => {
      tasks.value.push({
        id: resp.createTask, 
        name: 'Unnamed Task',
        body: 'New Task Body',
        tags: [],
      })
    })
}

const deleteTask = (taskId: string) => {
  $graphql.deleteTask({ taskId }).then(() => tasks.value = tasks.value.filter((t) => t.id !== taskId))
}
</script>

<template>
  <div>
    <div class="header">
      <h1>{{ me.name ? me.name : "My" }} To-Do List</h1>
      <button @click="logOut">
        Sign Out
      </button>
    </div>
    <button @click="refreshTasks">
      Refresh Tasks
    </button>
    <button @click="addTask">
      Add Task
    </button>
    <!-- We can't directly v-model task because `v-model cannot be used on v-for or v-slot scope variables because they are not writable.` -->
    <TaskRow
      v-for="(task, i) in tasks"
      :key="task.id"
      v-model:task="tasks[i]"
      @deleted="() => deleteTask(task.id)"
    />
  </div>
</template>

<style scoped>
button {
  margin: 0.25rem 0.5rem;
}

.header {
  display: flex;
  width: 100%;
  justify-content: space-between;
  align-items: center;
}
</style>