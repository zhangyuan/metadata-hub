<script setup lang="ts">
import { httpClient } from "@/http/client";
import { ref, onMounted } from "vue";

const datasetNames = ref<string[]>([]);

onMounted(async () => {
  const { data } = await httpClient.get("/api/datasets")
  datasetNames.value = data.data
})

</script>

<template>
  <div>
    <h1>databases</h1>

    <div>
      <ul>
        <li v-for="datasetName in datasetNames" v-bind:key="datasetName">
          <router-link :to="{name: 'dataset', params: {datasetName: datasetName}}">{{ datasetName }}</router-link>
        </li>
      </ul>
    </div>
  </div>

</template>

<style scoped></style>
