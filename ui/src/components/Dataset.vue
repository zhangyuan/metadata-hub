<script setup lang="ts">
import { httpClient } from "@/http/client";
import { defineProps, ref, onMounted } from "vue";


const tableNames = ref<string[]>([])

const props = defineProps({
  datasetName: String
})

onMounted(async () => {
  const { data } = await httpClient.get(`/api/datasets/${props.datasetName}`)
  tableNames.value = data.data.tables
})

</script>

<template>
  <div>
    <div class="breadcrumbs">
      <router-link :to="{name: 'datasets'}"> Databases </router-link>
      &gt;
      <span>{{ props.datasetName }}</span>
    </div>

    <h1>Dataset: {{ props.datasetName }}</h1>

    <div>
      <ul>
        <li v-for="tableName in tableNames" v-bind:key="tableName">
          <router-link :to="{name: 'table', params: {tableName: tableName}}">{{ tableName }}</router-link>
        </li>
      </ul>
    </div>
  </div>

</template>

<style scoped></style>
