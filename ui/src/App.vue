<script setup lang="ts">
import { RouterView } from "vue-router";
</script>

<template>
    <div>
        <div class="navbar">
            <ul>
                <li><router-link to="/">Home</router-link></li>
                <li><router-link to="/datasets">Datasets</router-link></li>
            </ul>
        </div>
        <RouterView />
    </div>
</template>

<style scoped></style>
