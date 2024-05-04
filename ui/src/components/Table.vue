<script setup lang="ts">
import { httpClient } from "@/http/client";
import { ref, onMounted } from "vue";


interface TableColumn {
  name: string,
  comments: string,
  type: string
}

interface Table {
  name: string,
  columns: TableColumn[]
}

interface TableResponse {
  data: Table
}

const props = defineProps(['datasetName', 'tableName'])

const table = ref<Table>()

onMounted(async() => {
  const { data } = await httpClient.get<TableResponse>(`/api/datasets/${props.datasetName}/tables/${props.tableName}`)
    table.value = data.data
})

</script>

<template>
  <div class="container mx-auto" v-if="table">
    <div class="breadcrumbs">
      <router-link to="/"> Home </router-link>
      &gt;
      <router-link :to="{name: 'datasets'}"> Datasets </router-link>
      &gt;
      <router-link :to="{name: 'dataset', params: {datasetName: props.datasetName}}">{{ props.datasetName }}</router-link>
      &gt;
      <span>{{ table?.name }}</span>
    </div>


    <h1 class="text-4xl my-4">Table: {{ table?.name }}</h1>

    <div>
      <table class="table-auto border-collapse border border-slate-400 my-5">
        <thead class="bg-cyan-200">
          <tr>
            <th class="p-2 border">Column</th>
            <th class="p-2 border">Type</th>
            <th class="p-2 border">Comments</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="column in table.columns" v-bind:key="column.name">
            <td class="p-2 border">{{ column.name }}</td>
            <td class="p-2 border">{{ column.type }}</td>
            <td class="p-2 border">{{ column.comments }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

</template>

<style scoped></style>
