import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import App from "@/App.vue";
import DatabasesList from "@/components/DatasetList.vue";
import Dataset from "./components/Dataset.vue";
import Table from "./components/Table.vue";
import { RouteLocation } from 'vue-router';

const routes = [{
  name: "root",
  path: "/",
  children: [
    {
      name: 'datasets',
      path: "/datasets",
      component: DatabasesList
    },
    {
      name: 'dataset',
      path: "/datasets/:datasetName",
      component: Dataset,
      props: (route: RouteLocation) => ({ datasetName: route.params.datasetName })
    },
    {
      name: 'table',
      path: "/datasets/:datasetName/tables/:tableName",
      component: Table,
      props: (route: RouteLocation) => ({
        datasetName: route.params.datasetName,
        tableName: route.params.tableName
      })
    }
  ]
}];

const app = createApp(App);

const router = createRouter({
  history: createWebHistory(),
  routes,
});
app.use(router);

app.mount("#app");
