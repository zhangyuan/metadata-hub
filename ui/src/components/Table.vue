<script setup lang="ts">
import { httpClient } from "@/http/client";
import { ref, onMounted, defineProps } from "vue";


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
  <div v-if="table">
    <div class="breadcrumbs">
      <router-link :to="{name: 'datasets'}"> Databases </router-link>
      &gt;
      <router-link :to="{name: 'dataset', params: {datasetName: props.datasetName}}">{{ props.datasetName }}</router-link>
      &gt;
      <span>{{ table?.name }}</span>
    </div>


    <h1>Table: {{ table?.name }}</h1>

    <div>
      <table>
        <thead>
          <tr>
            <th>Column</th>
            <th>Type</th>
            <th>Comments</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="column in table.columns" v-bind:key="column.name">
            <td>{{ column.name }}</td>
            <td>{{ column.type }}</td>
            <td>{{ column.comments }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

</template>

<style scoped></style>
