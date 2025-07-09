<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { connections, selectedConnection, testResults, loading } from '../store.js';
  import { GetConnections, CreateConnection, UpdateConnection, DeleteConnection, TestConnection } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let showCreateForm = false;
  let showEditForm = false;
  let showDeleteConfirm = false;
  
  let deletingConnection = null;
  let formConnection = {
    id: '',
    name: '',
    type: 'kafka',
    host: 'localhost',
    port: 9092,
    username: '',
    password: '',
    vhost: '',
    group_id: '',
    extra: null
  };

  let creating = false;
  let updating = false;
  let deleting = {};
  let testing = {};
  let initialLoadComplete = false;

  const defaultPorts = {
    kafka: 9092,
    rabbitmq: 5672,
    rocketmq: 9876,
  };

  const defaultHosts = {
    kafka: 'localhost',
    rabbitmq: 'localhost',
    rocketmq: '127.0.0.1'  // RocketMQ 使用 IP 地址避免解析问题
  };

  onMount(loadConnections);

  async function loadConnections() {
    loading.set(true);
    try {
      const newConnections = await GetConnections() || [];
      connections.set(newConnections);
    } catch (error) {
      dispatch('notification', { message: '加载连接失败: ' + error, type: 'error' });
    } finally {
      loading.set(false);
      initialLoadComplete = true;
    }
  }

  function openCreateForm() {
    formConnection = {
      id: `new_${Date.now()}`,
      name: '',
      type: 'kafka',
      host: defaultHosts.kafka,
      port: defaultPorts.kafka,
      username: '',
      password: '',
      vhost: '',
      group_id: '',
      extra: null
    };
    showCreateForm = true;
    showEditForm = false;
  }

  function openEditForm(connection) {
    // @ts-ignore
    formConnection = { ...connection };
    showEditForm = true;
    showCreateForm = false;
  }

  function handleCancel() {
    showCreateForm = false;
    showEditForm = false;
    formConnection = {
      id: '',
      name: '',
      type: 'kafka',
      host: 'localhost',
      port: 9092,
      username: '',
      password: '',
      vhost: '',
      group_id: '',
      extra: null
    };
  }

  function onTypeChange() {
    if (formConnection && formConnection.type in defaultPorts) {
        formConnection.port = defaultPorts[formConnection.type];
        formConnection.host = defaultHosts[formConnection.type];
        formConnection = {...formConnection};
    }
  }

  async function handleSubmit() {
    const isCreating = showCreateForm;
    if (isCreating) creating = true;
    else updating = true;

    try {
      const connData = {...formConnection};

      // 清理不需要的字段
      if(isCreating) {
        delete connData.id;
      }

      // 删除可能存在的时间字段，让后端自动处理
      if ('created' in connData) delete connData.created;
      if ('updated' in connData) delete connData.updated;

      if (isCreating) {
        // @ts-ignore
        await CreateConnection(connData);
        dispatch('notification', { message: '连接创建成功', type: 'success' });
      } else {
        // @ts-ignore
        await UpdateConnection(connData);
        dispatch('notification', { message: '连接更新成功', type: 'success' });
      }
      await loadConnections();
      handleCancel();
    } catch (error) {
      const message = isCreating ? '创建连接失败: ' : '更新连接失败: ';
      dispatch('notification', { message: message + error, type: 'error' });
    } finally {
      if (isCreating) creating = false;
      else updating = false;
    }
  }

  function startDelete(connection) {
    // @ts-ignore
    deletingConnection = connection;
    showDeleteConfirm = true;
  }

  function cancelDelete() {
    deletingConnection = null;
    showDeleteConfirm = false;
  }

  async function confirmDeleteConnection() {
    if (!deletingConnection) return;
    deleting[deletingConnection.id] = true;
    deleting = {...deleting};
    try {
      await DeleteConnection(deletingConnection.id);
      dispatch('notification', { message: `连接 "${deletingConnection.name}" 已删除`, type: 'success' });
      if ($selectedConnection && $selectedConnection.id === deletingConnection.id) {
        selectedConnection.set(null);
      }
      await loadConnections();
    } catch (error) {
      dispatch('notification', { message: '删除连接失败: ' + error, type: 'error' });
    } finally {
      deleting[deletingConnection.id] = false;
      deleting = {...deleting};
      cancelDelete();
    }
  }

  async function testConnection(connection) {
    // @ts-ignore
    testing[connection.id] = true;
    testing = { ...testing };
    try {
      const result = await TestConnection(connection.id);
      testResults.update(current => ({ ...current, [connection.id]: result }));

      const message = result.success
        ? `连接成功，延迟: ${result.latency}ms`
        : `连接失败: ${result.message}`;
      dispatch('notification', { message, type: result.success ? 'success' : 'error' });
    } catch (error) {
      const errorMessage = error.message || error.toString();
      testResults.update(current => ({ ...current, [connection.id]: { success: false, message: errorMessage } }));
      dispatch('notification', { message: `测试连接时发生错误: ${errorMessage}`, type: 'error' });
    } finally {
      testing[connection.id] = false;
      testing = { ...testing };
    }
  }



</script>

<div class="space-y-6">
  <!-- Connection Stats -->
  <div class="stats shadow w-full">
    <div class="stat">
        <div class="stat-figure text-primary">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
        </div>
        <div class="stat-title">总连接数</div>
        <div class="stat-value text-primary">{$connections.length}</div>
        <div class="stat-desc">管理您的所有消息队列连接</div>
    </div>
    <div class="stat">
        <div class="stat-figure text-secondary">
            {#if $selectedConnection && $testResults[$selectedConnection.id]}
                <div class={`badge ${$testResults[$selectedConnection.id].success ? 'badge-success' : 'badge-error'} gap-2`}>
                    {$testResults[$selectedConnection.id].success ? '在线' : '离线'}
                </div>
            {:else}
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="inline-block w-8 h-8 stroke-current"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            {/if}
        </div>
        <div class="stat-title">当前选中</div>
        <div class="stat-value text-secondary">{$selectedConnection ? $selectedConnection.name : '无'}</div>
        <div class="stat-desc">
            {#if $selectedConnection}
                {$selectedConnection.type} at {$selectedConnection.host}:{$selectedConnection.port}
            {:else}
                请选择一个连接以开始
            {/if}
        </div>
    </div>
  </div>

  <!-- Connection List -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center mb-4">
        <h2 class="card-title">连接列表</h2>
        <button class="btn btn-primary btn-sm" on:click={openCreateForm}>
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
          新建
        </button>
      </div>

      {#if $loading && !initialLoadComplete}
        <div class="text-center py-12"><span class="loading loading-spinner loading-lg"></span></div>
      {:else if $connections.length === 0}
        <div class="text-center py-12">
          <div class="text-6xl mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 inline-block text-base-content/30" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
            </svg>
          </div>
          <h3 class="text-lg font-semibold mb-2">暂无连接配置</h3>
          <p class="text-base-content/60 mb-4">开始创建您的第一个消息队列连接</p>
          <button class="btn btn-primary" on:click={openCreateForm}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
            创建连接
          </button>
        </div>
      {:else}
        <div class="space-y-3">
          {#each $connections as conn (conn.id)}
            <button 
              class="card card-compact bg-base-200/60 hover:bg-base-300/80 transition-all duration-200 ease-in-out cursor-pointer w-full text-left"
              class:ring-2={$selectedConnection && $selectedConnection.id === conn.id}
              class:ring-primary={$selectedConnection && $selectedConnection.id === conn.id}
              on:click={() => {
                selectedConnection.set(conn);
                // 自动测试连接状态
                if (!$testResults[conn.id]) {
                  testConnection(conn);
                }
              }}
            >
              <div class="card-body">
                <div class="flex justify-between items-center">
                  <div class="flex items-center gap-4">
                    <span class="relative flex h-3 w-3">
                      {#if testing[conn.id]}
                        <span class="absolute inline-flex h-full w-full rounded-full bg-info opacity-75 animate-ping"></span>
                      {/if}
                      <span class="relative inline-flex rounded-full h-3 w-3"
                            class:bg-info={testing[conn.id]}
                            class:bg-success={!testing[conn.id] && $testResults[conn.id]?.success}
                            class:bg-error={!testing[conn.id] && $testResults[conn.id] && !$testResults[conn.id].success}
                            class:opacity-30={!testing[conn.id] && !$testResults[conn.id]}
                            class:bg-base-content={!testing[conn.id] && !$testResults[conn.id]}></span>
                    </span>
                    <div>
                      <h3 class="card-title text-base">{conn.name}</h3>
                      <p class="text-xs text-base-content/70">{conn.type} - {conn.host}:{conn.port}</p>
                    </div>
                  </div>
                  <div class="card-actions">
                    <button class="btn btn-xs btn-ghost" on:click|stopPropagation={() => testConnection(conn)} disabled={testing[conn.id]}>
                      {#if testing[conn.id]}
                        <span class="loading loading-spinner loading-xs"></span>
                      {:else}
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
                      {/if}
                    </button>
                    <button class="btn btn-xs btn-ghost" on:click|stopPropagation={() => openEditForm(conn)}>
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.536L16.732 3.732z" /></svg>
                    </button>
                    <button class="btn btn-xs btn-ghost text-error" on:click|stopPropagation={() => startDelete(conn)} disabled={deleting[conn.id]}>
                      {#if deleting[conn.id]}
                        <span class="loading loading-spinner loading-xs"></span>
                      {:else}
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                      {/if}
                    </button>
                  </div>
                </div>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>

<!-- Create/Edit Connection Modal -->
{#if showCreateForm || showEditForm}
  <div class="modal modal-open">
    <div class="modal-box w-11/12 max-w-2xl">
      <h3 class="font-bold text-lg mb-4 flex items-center">
        {#if showCreateForm}
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
          新建连接
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.536L16.732 3.732z" /></svg>
          编辑连接
        {/if}
      </h3>
      
      <div class="form-control">
        <label for="conn-name-{formConnection.id}" class="label"><span class="label-text">连接名称</span></label>
        <input id="conn-name-{formConnection.id}" type="text" bind:value={formConnection.name} class="input input-bordered" />
      </div>

      <div class="form-control mt-4">
        <label for="conn-type-{formConnection.id}" class="label"><span class="label-text">消息队列类型</span></label>
        <select id="conn-type-{formConnection.id}" bind:value={formConnection.type} class="select select-bordered" on:change={onTypeChange}>
          <option value="kafka">Apache Kafka</option>
          <option value="rabbitmq">RabbitMQ</option>
          <option value="rocketmq">Apache RocketMQ</option>
        </select>
      </div>

      <div class="grid grid-cols-2 gap-4 mt-4">
        <div class="form-control">
          <label for="conn-host-{formConnection.id}" class="label">
            <span class="label-text">
              主机地址
              {#if formConnection.type === 'rocketmq'}
                <!-- <span class="text-warning">(推荐: 127.0.0.1)</span> -->
              {/if}
            </span>
          </label>
          <input
            id="conn-host-{formConnection.id}"
            type="text"
            bind:value={formConnection.host}
            class="input input-bordered"
            placeholder={formConnection.type === 'rocketmq' ? '127.0.0.1' : 'localhost'}
          />
        </div>
        <div class="form-control">
          <label for="conn-port-{formConnection.id}" class="label"><span class="label-text">端口</span></label>
          <input id="conn-port-{formConnection.id}" type="number" bind:value={formConnection.port} class="input input-bordered" />
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4 mt-4">
        <div class="form-control">
          <label for="conn-user-{formConnection.id}" class="label">
            <span class="label-text">
              用户名
              {#if formConnection.type === 'rabbitmq'}
                <span class="text-info">(默认: guest)</span>
              {:else}
                (可选)
              {/if}
            </span>
          </label>
          <input
            id="conn-user-{formConnection.id}"
            type="text"
            bind:value={formConnection.username}
            class="input input-bordered"
            placeholder={formConnection.type === 'rabbitmq' ? 'guest' : ''}
          />
        </div>
        <div class="form-control">
          <label for="conn-pass-{formConnection.id}" class="label">
            <span class="label-text">
              密码
              {#if formConnection.type === 'rabbitmq'}
                <span class="text-info">(默认: guest)</span>
              {:else}
                (可选)
              {/if}
            </span>
          </label>
          <input
            id="conn-pass-{formConnection.id}"
            type="password"
            bind:value={formConnection.password}
            class="input input-bordered"
            placeholder={formConnection.type === 'rabbitmq' ? 'guest' : ''}
          />
        </div>
      </div>

      <div class="modal-action">
        <button class="btn btn-primary" on:click={handleSubmit} disabled={creating || updating}>
          {#if creating || updating}
            <span class="loading loading-spinner"></span>
          {/if}
          {showCreateForm ? '创建' : '更新'}
        </button>
        <button class="btn btn-outline" on:click={handleCancel}>取消</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteConfirm}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg flex items-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
        确认删除
      </h3>
      <p class="py-4">确定要删除连接 "{deletingConnection.name}" 吗？</p>
      <div class="modal-action">
        <button class="btn btn-error" on:click={confirmDeleteConnection} disabled={deleting[deletingConnection.id]}>
          {#if deleting[deletingConnection.id]}
            <span class="loading loading-spinner"></span>
          {/if}
          删除
        </button>
        <button class="btn btn-outline" on:click={cancelDelete}>取消</button>
      </div>
    </div>
  </div>
{/if}