<script lang="ts">
  import { client } from '@arcanum/ts-client';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Password } from '$lib/components/ui/password';
  import AuthShell from '$lib/components/auth/AuthShell.svelte';

  let email = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  const login = async () => {
    loading = true;
    error = '';

    try {
      const res = await client.post('/auth/login', { email, password });
      const { token, spaces, current_space_id } = res.data;

      localStorage.setItem('arcanum_token', token);
      localStorage.setItem('arcanum_spaces', JSON.stringify(spaces));
      localStorage.setItem('arcanum_current_space', current_space_id);

      client.defaults.headers.common['Authorization'] = `Bearer ${token}`;

      window.location.href = '/panel';
    } catch (e: any) {
      error = e.response?.data?.message || 'Ошибка входа';
    } finally {
      loading = false;
    }
  };
</script>

<AuthShell title="Arcanum" subtitle="Войти в панель">
  <form on:submit|preventDefault={login} class="space-y-4">
    <div>
      <label class="text-sm font-medium">Email</label>
      <Input bind:value={email} placeholder="admin@example.com" required />
    </div>
    <div>
      <label class="text-sm font-medium">Пароль</label>
      <Password bind:value={password} placeholder="Пароль" required />
    </div>

    <Button type="submit" class="w-full" disabled={loading}>
      {loading ? 'Вход...' : 'Войти'}
    </Button>

    {#if error}
      <p class="text-red-500 text-sm">{error}</p>
    {/if}
  </form>
</AuthShell>