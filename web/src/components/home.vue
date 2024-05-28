<script setup>
import { inject } from 'vue';
import { useRouter } from 'vue-router';

const parameter = inject('parameters');
const router = useRouter();

if (parameter.value.defaultview && parameter.value.defaultview != "") {
    router.push(`/view/${parameter.value.defaultview}`);
}
</script>

<template>
    <div class="row">
        <div class="col-md-12">
            <div class="card" aria-hidden="true" style="position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);">
                <div class="card-header bg-primary">{{ $t('home.welcome') }}</div>
                <div class="card-body">
                    <h5 class="card-title placeholder-glow">
                        <span>{{ $t('home.header') }}</span>
                    </h5>
                    <p class="card-text">
                        {{ $t('home.message') }}
                    </p>
                </div>
            </div>
        </div>
    </div>
</template>
