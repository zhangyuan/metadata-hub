<script setup lang="ts">
import debounce from 'debounce';
import { httpClient } from "@/http/client";
import { ref, onMounted, watch } from "vue";
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';

interface ColumnHitFields {
  name: string
  comments: string,
  datasetName: string,
  tableName: string
}

interface ColumnHit {
  id: string,
  fields: ColumnHitFields
}

interface SearchColumnsResponseData {
  hits: ColumnHit[]
  total_hits: number
}

interface SearchColumnsResponse {
  data: SearchColumnsResponseData
}

interface TableHitFields {
  name: string
  datasetName: string
}

interface TableHit {
  id: string,
  fields: TableHitFields
}

interface SearchTablesResponseData {
  hits: TableHit[]
  total_hits: number
}

interface SearchTablesResponse {
  data: SearchTablesResponseData
}

const router = useRouter();
const route = useRoute();

const userInput = ref<string>("");

const searchColumnsResponse = ref<SearchColumnsResponse | null>()
const searchTablesResponse = ref<SearchTablesResponse | null>()

onMounted(() => {
  if (route.query.query) {
    userInput.value = route.query.query.toString()
  }
})

const updatePath = () => {
  router.push({ name: 'search', query: { query: userInput.value } })
}

const onSearch = updatePath

const searchColumns = debounce(async (q) => {
  const { data } = await httpClient.get<SearchColumnsResponse>('/api/search-columns', { params: { q: q } })
  searchColumnsResponse.value = data
}, 1000)

const searchTables = debounce(async (q) => {
  const { data } = await httpClient.get<SearchTablesResponse>('/api/search-tables', { params: { q: q } })
  searchTablesResponse.value = data
}, 1000)

watch(userInput, (value: string) => {
  if (value?.trim().length == 0) {
    searchColumnsResponse.value = null
    searchTablesResponse.value = null
  } else {
    searchColumns(value)
    searchTables(value)
  }
})

</script>

<template>
  <div class="container mx-auto">
    <h1 class="text-4xl text-center my-3">Search</h1>

    <div class="my-5 flex justify-center">
      <input class="border p-4 w-3/6" type="text" v-model="userInput" @input="onSearch">
    </div>

    <div class="">
      <div class="" v-if="searchTablesResponse">
        <h2 class="text-center"> Tables ({{ searchTablesResponse?.data.total_hits }}) </h2>

        <table class="table-auto w-full border-collapse border border-slate-400 my-5"
          v-if="searchTablesResponse?.data.hits.length > 0">
          <thead class="bg-cyan-200">
            <tr>
              <th class="p-2 border">Dataset</th>
              <th class="p-2 border">Table</th>
            </tr>
          </thead>

          <tbody>
            <tr class="hover:bg-lime-100" v-for="hit in searchTablesResponse?.data.hits" v-bind:key="hit.id">
              <td class="p-2 border">
                <router-link :to="{ name: 'dataset', params: { datasetName: hit.fields.datasetName } }">{{
                  hit.fields.datasetName }}</router-link>
              </td>
              <td class="p-2 border">
                <router-link
                  :to="{ name: 'table', params: { datasetName: hit.fields.datasetName, tableName: hit.fields.name } }">{{
                    hit.fields.name }}</router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="text-center" v-if="searchColumnsResponse">
        <h2> Columns ({{ searchColumnsResponse?.data.total_hits }}) </h2>

        <table class="table-auto w-full border-collapse border border-slate-400 my-5"
          v-if="searchColumnsResponse?.data.hits.length > 0">
          <thead class="bg-cyan-200">
            <tr>
              <th class="p-2 border">Dataset</th>
              <th class="p-2 border">Table</th>
              <th class="p-2 border">Column</th>
              <th class="p-2 border">Comments</th>
            </tr>
          </thead>

          <tbody>
            <tr class="hover:bg-lime-100" v-for="hit in searchColumnsResponse?.data.hits" v-bind:key="hit.id">
              <td class="p-2 border">
                <router-link :to="{ name: 'dataset', params: { datasetName: hit.fields.datasetName } }">{{
                  hit.fields.datasetName }}</router-link>
              </td>
              <td class="p-2 border">
                <router-link
                  :to="{ name: 'table', params: { datasetName: hit.fields.datasetName, tableName: hit.fields.tableName } }">{{
                    hit.fields.tableName }}</router-link>
              </td>
              <td class="p-2 border">{{ hit.fields.name }}</td>
              <td class="p-2 border">{{ hit.fields.comments }}</td>
            </tr>
          </tbody>
        </table>

      </div>
    </div>
  </div>

</template>

<style scoped></style>
