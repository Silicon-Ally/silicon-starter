<script setup lang="ts">
import { computed } from 'vue'
import { Task } from '@/graphql/generated'

const { $graphql } = useAPI()

interface Props {
  task: Task
}
const props = defineProps<Props>()

const emit = defineEmits<{
    (e: 'update:task', task: Task): void,
    (e: 'deleted'): void,
}>()

const prefix = `Task[${props.task.id}`
const isEditing = useState<boolean>(`${prefix}.isEditing`, () => false)
const isExpanded = useState<boolean>(`${prefix}.isExpanded`, () => false)

const cloneTask = (t: Task): Task => {
  return {
    id: t.id,
    name: t.name,
    body: t.body,
    tags: [...t.tags],
  }
}

const taskName = computed({
  get: () => props.task.name ?? 'Untitled Task',
  set: (newName: string) => {
    const cloned = cloneTask(props.task)
    cloned.name = newName
    emit('update:task', cloned)
  },
})
const taskBody = computed({
  get: () => props.task.body,
  set: (newBody: string) => {
    const cloned = cloneTask(props.task)
    cloned.body = newBody
    emit('update:task', cloned)
  },
})
const updateTaskName = () => $graphql.setTaskName({ taskId: props.task.id, newName: taskName.value })
const updateTaskBody = () => $graphql.setTaskBody({ taskId: props.task.id, newBody: taskBody.value })
const deleteTask = () => emit('deleted') 
</script>

<template>
  <div class="task">
    <div class="task-header">
      <input
        type="checkbox"
        @change="deleteTask"
      >
      <input
        v-if="isEditing"
        v-model="taskName"
        type="text"
        placeholder="Task Name..."
        @change="updateTaskName"
      >
      <span
        v-else
      >
        {{ taskName }}
      </span>
      <button @click="() => isEditing = !isEditing">
        {{ isEditing ? 'Done Editing' : 'Edit' }}
      </button>
      <button @click="() => isExpanded= !isExpanded">
        {{ isExpanded ? 'Hide Details' : 'Show Details' }}
      </button>
    </div>
    <div
      v-if="isExpanded"
    >
      <textarea
        v-if="isEditing"
        v-model="taskBody"
        placeholder="Task Body..."
        @change="updateTaskBody"
      />
      <span
        v-else
        @click="() => isExpanded = !isExpanded"
      >
        {{ taskBody }}
      </span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.task {
  display: flex;
  flex-direction: column;
  width: calc(100% - 1rem);
  border: 1px solid lightgray;
  border-radius: 0.5rem;
  padding: 0.25rem;
  margin: 0.5rem;
}

.task-header {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  width: 100%;
}

button {
  margin: 0.25rem;
}
</style>
