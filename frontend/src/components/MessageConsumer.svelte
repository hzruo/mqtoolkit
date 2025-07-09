<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { selectedConnection, selectedConsumerTopics, consumerState, consumerMessages } from '../store.js';
  import { StartConsuming, StopConsuming, ListTopics, SaveFile } from '../../wailsjs/go/main/App';
  import { eventManager } from '../eventManager.js';

  export let isOnline;
  const dispatch = createEventDispatcher();

  // 使用持久化的消费状态
  $: consuming = $consumerState.consuming;
  $: subscriptionId = $consumerState.subscriptionId;
  $: messages = $consumerMessages;
  let consumerConfig = {
    topics: '',
    groupId: 'mq-toolkit-consumer',
    fromBeginning: false,
    maxMessages: 100
  };

  let messageFilter = '';
  let availableTopics = [];
  let selectedTopics = [];
  let showTopicDropdown = false;

  // 通知函数
  function showNotification(message, type) {
    dispatch('notification', { message, type });
  }

  onMount(() => {
    console.log('MessageConsumer component mounted');

    // 设置全局事件监听器
    eventManager.setupEventListeners(showNotification, stopConsuming);

    // 恢复持久化的消费配置
    if ($consumerState.groupId) {
      consumerConfig.groupId = $consumerState.groupId;
    }
    if ($consumerState.fromBeginning !== undefined) {
      consumerConfig.fromBeginning = $consumerState.fromBeginning;
    }
    if ($consumerState.maxMessages) {
      consumerConfig.maxMessages = $consumerState.maxMessages;
    }

    // 返回清理函数 - 不清理事件监听器，让消费在后台继续
    return () => {
      console.log('Component unmounting, but keeping event listeners for background consumption');
      // 注意：这里不调用EventsOff，让事件监听器保持活跃
      // 这样即使页面切换，消费也能继续工作
    };
  });

  // 当连接改变时重新加载主题
  $: if ($selectedConnection) {
    loadTopics();
  }

  // 绑定主题到持久化存储
  $: consumerConfig.topics = $selectedConsumerTopics;
  $: selectedConsumerTopics.set(consumerConfig.topics);

  // 同步selectedTopics和consumerConfig.topics
  $: {
    if (consumerConfig.topics) {
      selectedTopics = consumerConfig.topics.split(',').map(t => t.trim()).filter(t => t);
    } else {
      selectedTopics = [];
    }
  }

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

  async function startConsuming() {
    if (!$selectedConnection || !isOnline) {
      dispatch('notification', { message: '请先选择一个在线的连接', type: 'error' });
      return;
    }

    const topics = consumerConfig.topics.split(',').map(t => t.trim()).filter(t => t);
    if (topics.length === 0) {
      dispatch('notification', { message: '请填写有效的消费主题', type: 'error' });
      return;
    }

    // 检查是否为Kafka且选择了多个主题
    if ($selectedConnection.type === 'kafka' && topics.length > 1) {
      dispatch('notification', { message: 'Kafka消费者只支持单个主题，请选择一个主题', type: 'error' });
      return;
    }

    try {
      consumerMessages.set([]);

      const req = {
        connection_id: $selectedConnection.id,
        topics: topics,
        group_id: consumerConfig.groupId,
        auto_commit: true,
        from_beginning: consumerConfig.fromBeginning,
      };

      console.log('Starting consumer with request:', req);
      const subId = await StartConsuming(req);

      // 更新持久化状态
      consumerState.update(state => ({
        ...state,
        consuming: true,
        subscriptionId: subId,
        groupId: consumerConfig.groupId,
        fromBeginning: consumerConfig.fromBeginning,
        maxMessages: consumerConfig.maxMessages
      }));

      console.log('Consumer started with subscription ID:', subId);
      dispatch('notification', { message: '已成功启动消费者', type: 'info' });

    } catch (error) {
      console.error('Failed to start consumer:', error);
      dispatch('notification', { message: `启动消费失败: ${error}`, type: 'error' });

      // 更新持久化状态
      consumerState.update(state => ({
        ...state,
        consuming: false,
        subscriptionId: null
      }));
    }
  }

  async function stopConsuming() {
    if (!subscriptionId) return;
    try {
      await StopConsuming(subscriptionId);
      dispatch('notification', { message: '已停止消费', type: 'info' });
    } catch (error) {
      dispatch('notification', { message: `停止消费失败: ${error}`, type: 'error' });
    } finally {
      // 更新持久化状态
      consumerState.update(state => ({
        ...state,
        consuming: false,
        subscriptionId: null
      }));
    }
  }

  function clearMessages() {
    consumerMessages.set([]);
  }

  async function loadTopics() {
    if (!$selectedConnection) return;
    try {
      const topics = await ListTopics($selectedConnection.id);
      availableTopics = topics.map(t => t.name);
    } catch (error) {
      console.error('Failed to load topics:', error);
      availableTopics = [];
    }
  }

  function toggleTopic(topic) {
    // 对于Kafka，只允许单选
    if ($selectedConnection && $selectedConnection.type === 'kafka') {
      selectedTopics = [topic];
    } else {
      // 对于RabbitMQ和RocketMQ，允许多选
      if (selectedTopics.includes(topic)) {
        selectedTopics = selectedTopics.filter(t => t !== topic);
      } else {
        selectedTopics = [...selectedTopics, topic];
      }
    }
    // 更新输入框
    consumerConfig.topics = selectedTopics.join(',');
  }

  $: filteredMessages = messages.filter(msg => {
    if (!messageFilter) return true;
    const filter = messageFilter.toLowerCase();
    const valueStr = typeof msg.value === 'string' ? msg.value : JSON.stringify(msg.value);
    return (msg.topic && msg.topic.toLowerCase().includes(filter)) ||
           (msg.key && msg.key.toLowerCase().includes(filter)) ||
           (valueStr.toLowerCase().includes(filter));
  });

  function formatJson(value) {
    try {
      const parsed = JSON.parse(value);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return value;
    }
  }

  function copyMessage(message) {
    const text = JSON.stringify(message, null, 2);
    navigator.clipboard.writeText(text).then(() => {
      dispatch('notification', { message: '消息已复制到剪贴板', type: 'success' });
    });
  }

  async function exportMessages() {
    try {
      const data = JSON.stringify(filteredMessages, null, 2);
      const filename = `messages-${new Date().toISOString().split('T')[0]}.json`;

      const savedPath = await SaveFile(filename, data);

      dispatch('notification', {
        message: `消息已导出 (${filteredMessages.length} 条消息)\n保存位置: ${savedPath}`,
        type: 'success'
      });

    } catch (error) {
      console.error('Export error:', error);
      if (error.message.includes('cancelled')) {
        dispatch('notification', { message: '导出已取消', type: 'info' });
      } else {
        dispatch('notification', { message: '导出失败: ' + error.message, type: 'error' });
      }
    }
  }
</script>

<div class="space-y-6">
  <!-- Connection Status -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body p-4">
      {#if $selectedConnection}
        <div class="flex items-center space-x-3">
          <div class="badge" class:badge-success={isOnline} class:badge-warning={!isOnline}>
            {isOnline ? '在线' : '离线'}
          </div>
          <span class="font-medium">{$selectedConnection.name}</span>
          <span class="text-sm text-base-content/60">({$selectedConnection.type.toUpperCase()})</span>
          {#if !isOnline}
            <span class="text-xs text-warning">连接未测试或已断开</span>
          {/if}
        </div>
      {:else}
        <div class="flex items-center space-x-3">
          <div class="badge badge-ghost">未连接</div>
          <span class="text-base-content/60">请先在连接管理页面选择一个连接</span>
        </div>
      {/if}
    </div>
  </div>

  <!-- Consumer Config -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center">
        <h2 class="card-title">消费配置</h2>
        {#if consuming}
          <div class="badge badge-success gap-2">
            <div class="w-2 h-2 bg-white rounded-full animate-pulse"></div>
            消费中
          </div>
        {:else}
          <div class="badge badge-outline">已停止</div>
        {/if}
      </div>
      <div class="form-control">
        <label for="consumer-topics" class="label">
          <span class="label-text">
            消费主题
            {#if $selectedConnection && $selectedConnection.type === 'kafka'}
              <span class="text-warning">(Kafka仅支持单主题)</span>
            {:else}
              (多个请用逗号隔开)
            {/if}
          </span>
          {#if availableTopics.length > 0}
            <div class="dropdown dropdown-end" class:dropdown-open={showTopicDropdown}>
              <div tabindex="0" role="button" class="btn btn-xs btn-outline"
                   on:click={() => showTopicDropdown = !showTopicDropdown}
                   on:keydown={(e) => e.key === 'Enter' && (showTopicDropdown = !showTopicDropdown)}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
                选择主题
              </div>
              <div class="dropdown-content z-[1] card card-compact w-72 p-0 shadow bg-base-100 border">
                <div class="card-body p-4">
                  <h3 class="card-title text-sm mb-3">可用主题 ({availableTopics.length})</h3>
                  <div class="space-y-2 max-h-48 overflow-y-auto">
                    {#each availableTopics as topic}
                      <label
                        class="label cursor-pointer justify-start p-2 rounded transition-colors"
                        class:opacity-60={isSystemTopic(topic)}
                        class:cursor-not-allowed={isSystemTopic(topic)}
                        class:hover:bg-base-200={!isSystemTopic(topic)}
                        class:bg-base-200={isSystemTopic(topic)}
                        title={isSystemTopic(topic) ? '系统主题不可选择' : ''}
                      >
                        <input
                          type={$selectedConnection && $selectedConnection.type === 'kafka' ? 'radio' : 'checkbox'}
                          class={$selectedConnection && $selectedConnection.type === 'kafka' ? 'radio radio-sm mr-3' : 'checkbox checkbox-sm mr-3'}
                          name={$selectedConnection && $selectedConnection.type === 'kafka' ? 'kafka-topic' : ''}
                          checked={selectedTopics.includes(topic)}
                          on:change={() => toggleTopic(topic)}
                          disabled={consuming || isSystemTopic(topic)}
                        />
                        <div class="flex items-center justify-between flex-1">
                          <div class="flex items-center">
                            {#if isSystemTopic(topic)}
                              <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-2 text-warning" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                              </svg>
                            {/if}
                            <span class="font-mono text-sm">{topic}</span>
                          </div>
                          {#if isSystemTopic(topic)}
                            <span class="badge badge-warning badge-xs ml-2 whitespace-nowrap">系统</span>
                          {/if}
                        </div>
                      </label>
                    {/each}
                  </div>
                  {#if selectedTopics.length > 0}
                    <div class="divider my-2"></div>
                    <div class="text-xs text-base-content/60">
                      已选择 {selectedTopics.length} 个主题
                      {#if $selectedConnection && $selectedConnection.type === 'kafka'}
                        <span class="text-warning">(Kafka仅支持单主题)</span>
                      {/if}
                    </div>
                  {/if}
                  <div class="mt-3">
                    <button class="btn btn-xs btn-outline w-full" on:click={() => showTopicDropdown = false}>
                      完成选择
                    </button>
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </label>
        <input id="consumer-topics" type="text" bind:value={consumerConfig.topics} class="input input-bordered w-full" placeholder="输入主题名称或从上方选择" disabled={!isOnline || consuming} />
      </div>
      <div class="form-control">
        <label for="consumer-group" class="label"><span class="label-text">消费组 ID</span></label>
        <input id="consumer-group" type="text" bind:value={consumerConfig.groupId} class="input input-bordered w-full" disabled={!isOnline || consuming} />
      </div>
      <div class="form-control">
        <label class="cursor-pointer label justify-start gap-4">
          <input type="checkbox" bind:checked={consumerConfig.fromBeginning} class="checkbox" disabled={!isOnline || consuming} />
          <span class="label-text">从最早的偏移量开始消费</span>
        </label>
      </div>
      <div class="card-actions justify-end">
        <button class="btn btn-primary" on:click={startConsuming} disabled={!isOnline || consuming}>
          {#if consuming}<span class="loading loading-spinner"></span>{/if}
          开始消费
        </button>
        <button class="btn btn-error" on:click={stopConsuming} disabled={!consuming}>停止消费</button>
      </div>
    </div>
  </div>

  <!-- Message List -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center mb-4">
        <h2 class="card-title">消息列表 ({filteredMessages.length})</h2>
        <div class="flex space-x-2">
          <button class="btn btn-sm btn-outline" on:click={clearMessages} disabled={messages.length === 0}>清空</button>
          <button class="btn btn-sm btn-outline" on:click={exportMessages} disabled={filteredMessages.length === 0}>导出</button>
        </div>
      </div>

      {#if consuming}
        <div class="alert alert-info mb-4">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
          <span>正在消费消息中...</span>
        </div>
      {/if}

      {#if messages.length > 0}
        <div class="form-control mb-4">
          <input type="text" bind:value={messageFilter} placeholder="过滤消息..." class="input input-bordered w-full" />
        </div>
      {/if}

      {#if filteredMessages.length === 0}
        <div class="text-center py-12">
          <div class="text-6xl mb-4">📭</div>
          <h3 class="text-lg font-semibold mb-2">暂无消息</h3>
          <p class="text-base-content/60">
            {#if consuming}
              等待消息到达...
            {:else}
              点击"开始消费"来接收消息
            {/if}
          </p>
        </div>
      {:else}
        <div class="space-y-4 max-h-96 overflow-y-auto">
          {#each filteredMessages as message, index (message.id || `msg-${index}-${message.timestamp || Date.now()}`)}
            <div class="card bg-base-200 compact">
              <div class="card-body p-4">
                <div class="flex justify-between items-start mb-2">
                  <div class="flex items-center space-x-2">
                    <span class="badge badge-primary">#{index + 1}</span>
                    <span class="font-mono text-sm">{message.topic}</span>
                    {#if message.key}
                      <span class="badge badge-outline">{message.key}</span>
                    {/if}
                  </div>
                  <div class="flex items-center space-x-2">
                    <span class="text-xs text-base-content/60">
                      {new Date(message.timestamp).toLocaleString()}
                    </span>
                    <button class="btn btn-xs btn-ghost" on:click={() => copyMessage(message)}>
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
                    </button>
                  </div>
                </div>
                <div class="bg-base-100 p-3 rounded font-mono text-sm whitespace-pre-wrap break-all">
                  {message.value}
                </div>
                {#if message.headers && Object.keys(message.headers).length > 0}
                  <div class="mt-2">
                    <span class="text-xs text-base-content/60">Headers:</span>
                    <div class="flex flex-wrap gap-1 mt-1">
                      {#each Object.entries(message.headers) as [key, value]}
                        <span class="badge badge-xs badge-outline">{key}: {value}</span>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>
