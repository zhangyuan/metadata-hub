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
  <div class="container mx-auto">
    <div class="breadcrumbs">
      <router-link to="/"> Home </router-link>
      &gt;
      <router-link :to="{name: 'datasets'}"> Datasets </router-link>
    </div>

    <h1 class="text-4xl my-4">Datasets</h1>

    <div class="my-5">
      <ul>
        <li class="bg-amber-200 my-4 p-4 hover:bg-amber-400" v-for="datasetName in datasetNames" v-bind:key="datasetName">
          <router-link :to="{name: 'dataset', params: {datasetName: datasetName}}">{{ datasetName }}</router-link>
        </li>
      </ul>
    </div>
  </div>

</template>

<style scoped></style>
