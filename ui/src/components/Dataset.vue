<script setup lang="ts">
import { httpClient } from "@/http/client";
import { ref, onMounted } from "vue";

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
  <div class="container mx-auto">
    <div class="breadcrumbs">
      <router-link to="/"> Home </router-link>
      &gt;
      <router-link :to="{name: 'datasets'}"> Datasets </router-link>
      &gt;
      <span>{{ props.datasetName }}</span>
    </div>

    <h1 class="text-4xl my-4">Dataset: {{ props.datasetName }}</h1>

    <div class="my-5">
      <ul>
        <li class="p-4 my-4 bg-lime-100 hover:bg-lime-400" v-for="tableName in tableNames" v-bind:key="tableName">
          <router-link :to="{name: 'table', params: {tableName: tableName}}">{{ tableName }}</router-link>
        </li>
      </ul>
    </div>
  </div>

</template>

<style scoped></style>
