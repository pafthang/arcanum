<script lang="ts">
  import { client } from '@arcanum/ts-client';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Password } from '$lib/components/ui/password';
  import AuthShell from '$lib/components/auth/AuthShell.svelte';

  let email = $state('');
  let password = $state('');
  let confirm = $state('');
  let loading = $state(false);
  let error = $state('');

  const register = async () => {
    loading = true;
    error = '';

    if (password !== confirm) {
      error = 'Пароли не совпадают';
      loading = false;
      return;
    }

    try {
      const res = await client.post('/auth/register', { email, password });
      const { token } = res.data;

      localStorage.setItem('arcanum_token', token);
      client.defaults.headers.common['Authorization'] = `Bearer ${token}`;

      window.location.href = '/panel';
    } catch (e: any) {
      error = e.response?.data?.message || 'Ошибка регистрации';
    } finally {
      loading = false;
    }
  };
</script>

<AuthShell title="Arcanum" subtitle="Создать первого администратора">
  <form on:submit|preventDefault={register} class="space-y-4">
    <div>
      <label class="text-sm font-medium">Email</label>
      <Input bind:value={email} placeholder="admin@example.com" required />
    </div>

    <div>
      <label class="text-sm font-medium">Пароль</label>
      <Password bind:value={password} placeholder="Мощный пароль" required />
    </div>

    <div>
      <label class="text-sm font-medium">Подтверждение</label>
      <Password bind:value={confirm} placeholder="Повторите пароль" required />
    </div>

    <Button type="submit" class="w-full" disabled={loading}>
      {loading ? 'Создание...' : 'Создать администратора'}
    </Button>

    {#if error}
      <p class="text-red-500 text-sm mt-2">{error}</p>
    {/if}
  </form>
</AuthShell>