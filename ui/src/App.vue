<script setup lang="ts">
import { RouterView } from "vue-router";
</script>

<template>
    <div class="">
        <div class="container mx-auto py-4">
            <ul class="flex justify-center">
                <li class="py-3 mx-2"><router-link to="/">Home</router-link></li>
                <li class="py-3 mx-2"><router-link to="/datasets">Datasets</router-link></li>
            </ul>
        </div>
        <RouterView />
    </div>
</template>

<style scoped></style>
