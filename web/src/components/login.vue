<script setup>
import { ref, inject } from 'vue';
import axios from 'axios';

import { useRouter } from 'vue-router'; // Ajout de l'importation du routeur

const router = useRouter(); // Initialisation du routeur

const username = ref('');
const password = ref('');

const apiUrl = inject('apiUrl');
const errorMessage = ref(null);

const login = async () => {
    try {
        const response = await axios.post(`${apiUrl}/auth/login`, {
            username: username.value,
            password: password.value
        });

        // Rediriger vers la page appropriée après la connexion
        router.push("/redirect");
    } catch (error) {
        console.error(error);
        // Afficher un message d'erreur à l'utilisateur
        errorMessage.value = 'Authentication failed. Please check your username and password.';
    }
};
</script>

<template>
    <!-- Login 13 - Bootstrap Brain Component -->
    <section class="py-3 py-md-5">
        <div class="container">
            <div class="row justify-content-center">
                <div class="col-12 col-sm-10 col-md-8 col-lg-6 col-xl-5 col-xxl-4">
                    <div class="card border border-light-subtle rounded-3 shadow-sm">
                        <div class="card-body p-3 p-md-4 p-xl-5">
                            <div class="text-center mb-3">
                                <a href="#!">
                                    <img src="/logo.png" alt="BootstrapBrain Logo" width="175">
                                </a>
                            </div>
                            <h2 class="fs-6 fw-normal text-center text-secondary mb-4">Sign in to your account</h2>
                            <form @submit.prevent="login">
                                <div class="row gy-2 overflow-hidden">
                                    <div class="col-12">
                                        <div class="form-floating mb-3">
                                            <input type="text" class="form-control" name="username" id="username"
                                                placeholder="Username" v-model="username" required>
                                            <label for="username" class="form-label">Username</label>
                                        </div>
                                    </div>
                                    <div class="col-12">
                                        <div class="form-floating mb-3">
                                            <input type="password" class="form-control" name="password" id="password"
                                                value="" placeholder="Password" v-model="password" required>
                                            <label for="password" class="form-label">Password</label>
                                        </div>
                                    </div>
                                    <div class="col-12">
                                        <div class="d-grid my-3">
                                            <button class="btn btn-primary btn-lg" type="submit">Log in</button>
                                        </div>
                                    </div>
                                </div>
                            </form>
                            <!-- Afficher le message d'erreur s'il y a une erreur -->
                            <div v-if="errorMessage" class="alert alert-danger mt-3" role="alert">
                                {{ errorMessage }}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>