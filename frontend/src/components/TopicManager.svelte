<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { selectedConnection } from '../store.js';
  import { ListTopics, CreateTopic, DeleteTopic } from '../../wailsjs/go/main/App.js';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';

  export let isOnline;
  const dispatch = createEventDispatcher();

  let topics = [];
  let loading = false;
  let showCreateForm = false;
  let newTopic = {
    name: '',
    partitions: 1,
    replicas: 1
  };
  let creating = false;
  let deleting = {};
  let showDeleteConfirm = false;
  let topicToDelete = null;

  // 判断是否为系统内部主题
  function isSystemTopic(topicName) {
    const systemTopics = [
      // Kafka系统主题
      '__consumer_offsets',
      '__transaction_state',
      '_schemas',
      '__confluent.support.metrics',
      '_confluent-metrics',
      '_confluent-command',
      '_confluent-monitoring',
      // RabbitMQ系统队列
      'amq.direct',
      'amq.fanout',
      'amq.topic',
      'amq.headers',
      'amq.match',
      'amq.rabbitmq.trace',
      'amq.rabbitmq.log',
      // RocketMQ系统主题
      'TBW102',
      'BenchmarkTest',
      'OFFSET_MOVED_EVENT',
      'DefaultCluster',
      'SELF_TEST_TOPIC',
      'RMQ_SYS_TRANS_HALF_TOPIC',
      'RMQ_SYS_TRACE_TOPIC',
      '%RETRY%',
      '%DLQ%'
    ];

    return systemTopics.some(pattern => {
      if (pattern.includes('%')) {
        // 处理包含通配符的模式
        const regex = new RegExp(pattern.replace(/%/g, '.*'));
        return regex.test(topicName);
      }
      return topicName === pattern || topicName.startsWith(pattern);
    });
  }

  async function loadTopics() {
    if (!$selectedConnection || !isOnline) {
      topics = [];
      return;
    }
    
    try {
      loading = true;
      const result = await ListTopics($selectedConnection.id);
      topics = result || [];
    } catch (error) {
      dispatch('notification', { message: '加载主题列表失败: ' + error, type: 'error' });
      topics = [];
    } finally {
      loading = false;
    }
  }

  $: if ($selectedConnection && isOnline) {
    loadTopics();
  } else {
    topics = [];
  }

  async function createTopic() {
    if (!newTopic.name) {
      dispatch('notification', { message: '请填写主题名称', type: 'error' });
      return;
    }
    if (!$selectedConnection) {
      dispatch('notification', { message: '没有选中的连接', type: 'error' });
      return;
    }

    try {
      creating = true;
      const req = {
        connection_id: $selectedConnection.id,
        topic: newTopic.name,
        partitions: newTopic.partitions,
        replicas: newTopic.replicas
      };
      await CreateTopic(req);
      dispatch('notification', { message: '主题创建成功', type: 'success' });
      showCreateForm = false;
      newTopic = { name: '', partitions: 1, replicas: 1 };
      await loadTopics();
    } catch (error) {
      dispatch('notification', { message: `创建主题失败: ${error}`, type: 'error' });
    } finally {
      creating = false;
    }
  }

  function startDeleteTopic(topicName) {
    topicToDelete = topicName;
    showDeleteConfirm = true;
  }

  function cancelDelete() {
    topicToDelete = null;
    showDeleteConfirm = false;
  }

  function openDashboard() {
    // 打开RocketMQ Dashboard
    BrowserOpenURL('http://localhost:8080');
  }

  async function confirmDelete() {
    if (!topicToDelete || !$selectedConnection) {
      dispatch('notification', { message: '没有选中的连接或主题', type: 'error' });
      return;
    }

    try {
      deleting[topicToDelete] = true;
      deleting = {...deleting};
      const req = {
        connection_id: $selectedConnection.id,
        topic: topicToDelete
      };
      await DeleteTopic(req);
      dispatch('notification', { message: '主题删除成功', type: 'success' });
      await loadTopics();
    } catch (error) {
      dispatch('notification', { message: `删除主题失败: ${error}`, type: 'error' });
    } finally {
      deleting[topicToDelete] = false;
      deleting = {...deleting};
      cancelDelete();
    }
  }

</script>

<div class="space-y-6">
  <!-- Connection Status & Controls -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center mb-4">
        <h2 class="card-title">主题/队列</h2>
        {#if $selectedConnection && isOnline}
          <button class="btn btn-primary btn-sm" on:click={() => showCreateForm = true}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" /></svg>
            新建主题
          </button>
        {/if}
      </div>

      {#if !$selectedConnection}
        <div class="alert alert-info">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          <span>请先在“连接管理”页面选择一个连接。</span>
        </div>
      {:else if !isOnline}
        <div class="alert alert-warning">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path></svg>
          <span>当前连接 <span class="font-bold">{$selectedConnection.name}</span> 不在线，请先测试连接。</span>
        </div>
      {/if}
    </div>
  </div>

  <!-- Topic List -->
  {#if $selectedConnection && isOnline}
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        {#if loading}
          <div class="text-center py-12">
            <span class="loading loading-spinner loading-lg"></span>
            <p class="mt-4 text-base-content/60">正在加载主题列表...</p>
          </div>
        {:else if topics.length === 0}
          <div class="text-center py-12">
            {#if $selectedConnection && $selectedConnection.type === 'rocketmq'}
              <div class="text-6xl mb-4">🚀</div>
              <h3 class="text-lg font-semibold mb-2">RocketMQ 主题列表</h3>
              <p class="text-base-content/60 mb-4">RocketMQ v2 admin API 暂不支持列出所有主题</p>
              <div class="alert alert-info">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                <div>
                  <div class="font-semibold">替代方案：</div>
                  <ul class="list-disc list-inside mt-2 text-sm">
                    <li>使用 <button class="link link-primary" on:click={() => openDashboard()}>RocketMQ Dashboard</button> 查看主题 (默认: localhost:8080)</li>
                    <li>在消息发送页面直接输入主题名称</li>
                    <li>创建新主题后可在 Dashboard 中查看</li>
                  </ul>
                </div>
              </div>
            {:else}
              <div class="text-6xl mb-4">📂</div>
              <h3 class="text-lg font-semibold mb-2">暂无主题</h3>
              <p class="text-base-content/60">此连接下没有找到任何主题，您可以新建一个。</p>
            {/if}
          </div>
        {:else}
          <div class="overflow-x-auto">
            <table class="table w-full">
              <thead>
                <tr>
                  <th>主题名称</th>
                  {#if $selectedConnection.type === 'kafka'}
                    <th>分区数</th>
                    <th>副本数</th>
                  {/if}
                  <th class="text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {#each topics as topic (topic.name)}
                  <tr class:opacity-70={isSystemTopic(topic.name)} class:bg-base-200={isSystemTopic(topic.name)}>
                    <td class="font-mono">
                      <div class="flex items-center justify-between">
                        <div class="flex items-center">
                          {#if isSystemTopic(topic.name)}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-warning" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                            </svg>
                          {/if}
                          <span>{topic.name}</span>
                        </div>
                        {#if isSystemTopic(topic.name)}
                          <span class="badge badge-warning badge-xs whitespace-nowrap">系统</span>
                        {/if}
                      </div>
                    </td>
                    {#if $selectedConnection.type === 'kafka'}
                      <td>{topic.partitions}</td>
                      <td>{topic.replicas}</td>
                    {/if}
                    <td class="text-right">
                      {#if isSystemTopic(topic.name)}
                        <div class="tooltip" data-tip="系统主题不可删除">
                          <button class="btn btn-xs btn-ghost text-base-content/30" disabled>
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                          </button>
                        </div>
                      {:else}
                        <button
                          class="btn btn-xs btn-ghost text-error"
                          on:click|stopPropagation={() => startDeleteTopic(topic.name)}
                          disabled={deleting[topic.name]}
                        >
                          {#if deleting[topic.name]}
                            <span class="loading loading-spinner loading-xs"></span>
                          {:else}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                          {/if}
                        </button>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<!-- Create Topic Modal -->
{#if showCreateForm}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg">新建主题</h3>
      <div class="py-4 space-y-4">
        <div class="form-control">
          <label for="topic-name" class="label"><span class="label-text">主题名称</span></label>
          <input id="topic-name" type="text" bind:value={newTopic.name} class="input input-bordered" />
        </div>
        {#if $selectedConnection.type === 'kafka'}
          <div class="form-control">
            <label for="topic-partitions" class="label"><span class="label-text">分区数</span></label>
            <input id="topic-partitions" type="number" bind:value={newTopic.partitions} class="input input-bordered" />
          </div>
          <div class="form-control">
            <label for="topic-replicas" class="label"><span class="label-text">副本数</span></label>
            <input id="topic-replicas" type="number" bind:value={newTopic.replicas} class="input input-bordered" />
          </div>
        {/if}
      </div>
      <div class="modal-action">
        <button class="btn btn-primary" on:click={createTopic} disabled={creating}>
          {#if creating}<span class="loading loading-spinner"></span>{/if}
          创建
        </button>
        <button class="btn btn-outline" on:click={() => showCreateForm = false} disabled={creating}>取消</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteConfirm}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg">确认删除</h3>
      <p class="py-4">确定要删除主题 <span class="font-mono bg-base-200 px-2 py-1 rounded">"{topicToDelete}"</span> 吗？</p>
      <p class="text-warning text-sm">⚠️ 此操作不可恢复，主题中的所有消息都将丢失。</p>
      <div class="modal-action">
        <button class="btn btn-error" on:click={confirmDelete}>确认删除</button>
        <button class="btn btn-outline" on:click={cancelDelete}>取消</button>
      </div>
    </div>
  </div>
{/if}