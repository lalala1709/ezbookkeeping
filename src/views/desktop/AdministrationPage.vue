<template>
    <div class="admin-page d-flex align-center justify-center pa-4">
        <v-card class="admin-card" elevation="3">
            <v-card-item>
                <template #prepend>
                    <v-avatar color="primary" variant="tonal">
                        <v-icon icon="mdi-shield-account-outline" />
                    </v-avatar>
                </template>
                <v-card-title>Quản trị thành viên</v-card-title>
                <v-card-subtitle>ezBookkeeping</v-card-subtitle>
                <template #append v-if="isAuthenticated">
                    <v-btn variant="text" size="small" :disabled="loading" @click="logout">
                        Đăng xuất quản trị
                    </v-btn>
                </template>
            </v-card-item>

            <v-card-text v-if="!isAuthenticated">
                <p class="text-body-1 mb-5">Nhập mật khẩu quản trị để xem và quản lý thành viên.</p>
                <v-text-field
                    v-model="loginPassword"
                    type="password"
                    autocomplete="current-password"
                    autofocus
                    hide-details="auto"
                    label="Mật khẩu quản trị"
                    :disabled="loading"
                    @keyup.enter="login"
                />
                <v-alert class="mt-4" density="compact" type="error" v-if="errorMessage">
                    {{ errorMessage }}
                </v-alert>
                <v-btn block class="mt-5" color="primary" :loading="loading" :disabled="!loginPassword" @click="login">
                    Mở trang quản trị
                </v-btn>
                <div class="text-caption text-medium-emphasis mt-5">
                    Mật khẩu này là mật khẩu riêng của máy chủ, không phải mật khẩu của một thành viên.
                </div>
            </v-card-text>

            <template v-else>
                <v-card-text>
                    <v-alert density="compact" type="info" variant="tonal">
                        Có <strong>{{ totalUserCount }}</strong> thành viên đang hoạt động. Thao tác xóa dữ liệu và xóa thành viên không thể hoàn tác.
                    </v-alert>
                    <v-alert class="mt-4" density="compact" type="error" v-if="errorMessage">
                        {{ errorMessage }}
                    </v-alert>
                    <v-alert class="mt-4" density="compact" type="success" v-if="successMessage">
                        {{ successMessage }}
                    </v-alert>
                </v-card-text>

                <div class="admin-table-wrap">
                    <v-table density="comfortable">
                        <thead>
                            <tr>
                                <th>Thành viên</th>
                                <th>Trạng thái</th>
                                <th>Đăng ký</th>
                                <th>Đăng nhập gần nhất</th>
                                <th class="text-right">Thao tác</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="user in users" :key="user.username">
                                <td>
                                    <div class="font-weight-medium">{{ user.nickname }}</div>
                                    <div class="text-body-2">{{ user.username }}</div>
                                    <div class="text-caption text-medium-emphasis">{{ user.email }}</div>
                                </td>
                                <td>
                                    <v-chip size="small" :color="user.disabled ? 'warning' : 'success'" variant="tonal">
                                        {{ user.disabled ? 'Đã khóa' : 'Hoạt động' }}
                                    </v-chip>
                                </td>
                                <td>{{ formatTime(user.createdAt) }}</td>
                                <td>{{ formatTime(user.lastLoginAt) }}</td>
                                <td class="text-right text-no-wrap">
                                    <v-btn class="me-1" size="small" variant="tonal" :disabled="actionBusy" @click="openPasswordDialog(user)">
                                        Đổi mật khẩu
                                    </v-btn>
                                    <v-btn class="me-1" color="warning" size="small" variant="tonal" :disabled="actionBusy" @click="openClearDataDialog(user)">
                                        Xóa dữ liệu
                                    </v-btn>
                                    <v-btn color="error" size="small" variant="tonal" :disabled="actionBusy" @click="openDeleteDialog(user)">
                                        Xóa thành viên
                                    </v-btn>
                                </td>
                            </tr>
                            <tr v-if="!loading && users.length === 0">
                                <td class="text-center text-medium-emphasis py-8" colspan="5">Chưa có thành viên nào.</td>
                            </tr>
                        </tbody>
                    </v-table>
                </div>

                <v-card-actions class="px-6 py-4">
                    <v-btn variant="text" :loading="loading" :disabled="actionBusy" @click="loadUsers">
                        Làm mới danh sách
                    </v-btn>
                    <v-spacer />
                    <span class="text-caption text-medium-emphasis">Mật khẩu quản trị chỉ được giữ trong phiên trang này.</span>
                </v-card-actions>
            </template>
        </v-card>

        <v-dialog v-model="showPasswordDialog" max-width="480" persistent>
            <v-card>
                <v-card-title>Đổi mật khẩu</v-card-title>
                <v-card-text>
                    <p>Đặt mật khẩu mới cho <strong>{{ selectedUser?.username }}</strong>. Tất cả phiên đăng nhập hiện tại của thành viên này sẽ bị đăng xuất.</p>
                    <v-text-field
                        v-model="newPassword"
                        type="password"
                        autocomplete="new-password"
                        label="Mật khẩu mới"
                        hint="Tối thiểu 6 ký tự"
                        :disabled="actionBusy"
                        @keyup.enter="updatePassword"
                    />
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="actionBusy" @click="showPasswordDialog = false">Hủy</v-btn>
                    <v-btn color="primary" :loading="actionBusy" :disabled="newPassword.length < 6" @click="updatePassword">Lưu mật khẩu</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>

        <v-dialog v-model="showClearDataDialog" max-width="520" persistent>
            <v-card>
                <v-card-title>Xóa dữ liệu thành viên?</v-card-title>
                <v-card-text>
                    <p>Dữ liệu tài chính của <strong>{{ selectedUser?.username }}</strong> sẽ bị xóa: giao dịch, tài khoản, danh mục, nhãn, mẫu giao dịch và dữ liệu phụ. Tài khoản đăng nhập vẫn được giữ lại.</p>
                    <p class="text-error font-weight-medium">Không thể hoàn tác.</p>
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="actionBusy" @click="showClearDataDialog = false">Hủy</v-btn>
                    <v-btn color="warning" :loading="actionBusy" @click="clearUserData">Xóa dữ liệu</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>

        <v-dialog v-model="showDeleteDialog" max-width="520" persistent>
            <v-card>
                <v-card-title>Xóa thành viên?</v-card-title>
                <v-card-text>
                    <p>Tài khoản <strong>{{ selectedUser?.username }}</strong> sẽ bị xóa và toàn bộ phiên đăng nhập của họ sẽ bị đăng xuất.</p>
                    <p class="text-error font-weight-medium">Để xác nhận, hãy nhập đúng tên đăng nhập bên dưới.</p>
                    <v-text-field
                        v-model="deleteConfirmation"
                        autocomplete="off"
                        label="Tên đăng nhập để xác nhận"
                        :disabled="actionBusy"
                        @keyup.enter="deleteUser"
                    />
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn :disabled="actionBusy" @click="showDeleteDialog = false">Hủy</v-btn>
                    <v-btn color="error" :loading="actionBusy" :disabled="deleteConfirmation !== selectedUser?.username" @click="deleteUser">Xóa thành viên</v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

interface AdminUser {
    username: string;
    email: string;
    nickname: string;
    disabled: boolean;
    emailVerified: boolean;
    createdAt: string;
    lastLoginAt: string;
}

interface AdminUserListResult {
    totalUserCount: string;
    users: AdminUser[];
}

interface ApiEnvelope<T> {
    success: boolean;
    result?: T;
    errorMessage?: string;
}

const loginPassword = ref<string>('');
const adminPassword = ref<string>('');
const isAuthenticated = ref<boolean>(false);
const loading = ref<boolean>(false);
const actionBusy = ref<boolean>(false);
const errorMessage = ref<string>('');
const successMessage = ref<string>('');
const totalUserCount = ref<string>('0');
const users = ref<AdminUser[]>([]);

const selectedUser = ref<AdminUser | null>(null);
const newPassword = ref<string>('');
const deleteConfirmation = ref<string>('');
const showPasswordDialog = ref<boolean>(false);
const showClearDataDialog = ref<boolean>(false);
const showDeleteDialog = ref<boolean>(false);

const selectedUsername = computed<string>(() => selectedUser.value?.username || '');

async function adminApi<T>(path: string, method = 'GET', requestData?: Record<string, string>): Promise<T> {
    const response = await fetch(`/api/admin/${path}`, {
        method,
        headers: {
            'Content-Type': 'application/json',
            'X-EZB-Admin-Password': adminPassword.value
        },
        body: requestData ? JSON.stringify(requestData) : undefined
    });

    const envelope = await response.json() as ApiEnvelope<T>;

    if (!response.ok || !envelope.success || typeof envelope.result === 'undefined') {
        if (response.status === 401) {
            isAuthenticated.value = false;
            adminPassword.value = '';
        }

        throw new Error(envelope.errorMessage || 'Không thể thực hiện yêu cầu quản trị.');
    }

    return envelope.result;
}

async function loadUsers(): Promise<void> {
    loading.value = true;
    errorMessage.value = '';

    try {
        const result = await adminApi<AdminUserListResult>('users.json');
        totalUserCount.value = result.totalUserCount;
        users.value = result.users;
    } catch (error) {
        errorMessage.value = getErrorMessage(error);
    } finally {
        loading.value = false;
    }
}

async function login(): Promise<void> {
    if (!loginPassword.value || loading.value) {
        return;
    }

    adminPassword.value = loginPassword.value;
    await loadUsers();

    if (!errorMessage.value) {
        loginPassword.value = '';
        isAuthenticated.value = true;
    }
}

function logout(): void {
    adminPassword.value = '';
    isAuthenticated.value = false;
    users.value = [];
    totalUserCount.value = '0';
    errorMessage.value = '';
    successMessage.value = '';
}

function openPasswordDialog(user: AdminUser): void {
    selectedUser.value = user;
    newPassword.value = '';
    showPasswordDialog.value = true;
}

function openClearDataDialog(user: AdminUser): void {
    selectedUser.value = user;
    showClearDataDialog.value = true;
}

function openDeleteDialog(user: AdminUser): void {
    selectedUser.value = user;
    deleteConfirmation.value = '';
    showDeleteDialog.value = true;
}

async function updatePassword(): Promise<void> {
    if (!selectedUsername.value || newPassword.value.length < 6) {
        return;
    }

    await runAction(async () => {
        await adminApi<boolean>('users/password.json', 'POST', {
            username: selectedUsername.value,
            password: newPassword.value
        });
        showPasswordDialog.value = false;
        successMessage.value = `Đã đổi mật khẩu và đăng xuất các phiên của ${selectedUsername.value}.`;
    });
}

async function clearUserData(): Promise<void> {
    if (!selectedUsername.value) {
        return;
    }

    await runAction(async () => {
        await adminApi<boolean>('users/clear_data.json', 'POST', { username: selectedUsername.value });
        showClearDataDialog.value = false;
        successMessage.value = `Đã xóa dữ liệu của ${selectedUsername.value}.`;
    });
}

async function deleteUser(): Promise<void> {
    if (!selectedUsername.value || deleteConfirmation.value !== selectedUsername.value) {
        return;
    }

    await runAction(async () => {
        await adminApi<boolean>('users/delete.json', 'POST', { username: selectedUsername.value });
        showDeleteDialog.value = false;
        successMessage.value = `Đã xóa thành viên ${selectedUsername.value}.`;
        await loadUsers();
    });
}

async function runAction(action: () => Promise<void>): Promise<void> {
    actionBusy.value = true;
    errorMessage.value = '';
    successMessage.value = '';

    try {
        await action();
    } catch (error) {
        errorMessage.value = getErrorMessage(error);
    } finally {
        actionBusy.value = false;
    }
}

function formatTime(value: string): string {
    const unixTime = Number(value);

    if (!unixTime) {
        return 'Chưa đăng nhập';
    }

    return new Intl.DateTimeFormat('vi-VN', {
        dateStyle: 'medium',
        timeStyle: 'short'
    }).format(new Date(unixTime * 1000));
}

function getErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message;
    }

    return 'Không thể thực hiện yêu cầu quản trị.';
}
</script>

<style scoped>
.admin-page {
    min-height: 100vh;
    background: rgb(var(--v-theme-background));
}

.admin-card {
    width: min(1180px, 100%);
}

.admin-table-wrap {
    overflow-x: auto;
}

.admin-table-wrap :deep(th),
.admin-table-wrap :deep(td) {
    white-space: nowrap;
}

.admin-table-wrap :deep(td:first-child) {
    min-width: 230px;
}
</style>
