<template>
    <Views>
        <template #title>Client Management</template>
        <v-card>
            <v-card-title>MSP Clients</v-card-title>
            <v-card-text>
                <v-list>
                    <v-list-item v-for="client in clients" :key="client.id">
                        <v-list-item-content>
                            <v-list-item-title>{{ client.name }}</v-list-item-title>
                            <v-list-item-subtitle>{{ client.email }}</v-list-item-subtitle>
                        </v-list-item-content>
                        <v-list-item-action>
                            <v-btn icon>
                                <v-icon>mdi-pencil</v-icon>
                            </v-btn>
                        </v-list-item-action>
                    </v-list-item>
                </v-list>
                <v-btn color="primary" @click="addClient">Add Client</v-btn>
            </v-card-text>
        </v-card>
    </Views>
</template>

<script>
import Views from '@/views/Views.vue';

export default {
    components: { Views },
    data() {
        return {
            clients: [
                { id: 1, name: 'Client A', email: 'clienta@example.com' },
                { id: 2, name: 'Client B', email: 'clientb@example.com' },
            ],
        };
    },
    methods: {
        addClient() {
            // Placeholder for adding client
            alert('Add client functionality');
        },
    },
};
</script>