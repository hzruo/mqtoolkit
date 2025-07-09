<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { selectedConnection, selectedProducerTopic } from '../store.js';
  import { ProduceMessage, ListTemplates, ListTopics } from '../../wailsjs/go/main/App.js';

  export let isOnline;
  const dispatch = createEventDispatcher();

  let message = {
    topic: '',
    key: '',
    value: '',
    headers: {}
  };

  let sending = false;
  let showAdvanced = false;
  let headerKey = '';
  let headerValue = '';
  let messageTemplates = [];
  let showTemplateModal = false;
  let availableTopics = [];
  let showTopicDropdown = false;

  async function loadTemplates() {
    try {
      messageTemplates = await ListTemplates() || [];
      console.log('Loaded templates:', messageTemplates);
    } catch (error) {
      console.error('Failed to load templates:', error);
      dispatch('notification', { message: '加载消息模板失败: ' + error, type: 'error' });
    }
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

  onMount(loadTemplates);

  // 当连接改变时重新加载主题
  $: if ($selectedConnection) {
    loadTopics();
  }

  // 绑定主题到持久化存储
  $: message.topic = $selectedProducerTopic;
  $: selectedProducerTopic.set(message.topic);

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

  async function sendMessage() {
    if (!$selectedConnection || !isOnline) {
      dispatch('notification', { message: '请先选择一个在线的连接', type: 'error' });
      return;
    }

    if (!message.topic || !message.value) {
      dispatch('notification', { message: '请填写主题和消息内容', type: 'error' });
      return;
    }

    try {
      sending = true;
      
      const request = {
        connection_id: $selectedConnection.id,
        topic: message.topic,
        key: message.key,
        value: message.value,
        headers: message.headers
      };

      await ProduceMessage(request);
      dispatch('notification', { message: '消息发送成功', type: 'success' });
      
      message.value = '';
      message.key = '';
      message.headers = {};
      
    } catch (error) {
      dispatch('notification', { message: '发送消息失败: ' + error, type: 'error' });
    } finally {
      sending = false;
    }
  }

  function addHeader() {
    if (headerKey && headerValue) {
      message.headers[headerKey] = headerValue;
      headerKey = '';
      headerValue = '';
      message.headers = message.headers;
    }
  }

  function removeHeader(key) {
    delete message.headers[key];
    message.headers = { ...message.headers };
  }

  function openTemplateModal() {
    showTemplateModal = true;
  }

  function closeTemplateModal() {
    showTemplateModal = false;
  }

  function useTemplate(template) {
    message.value = template.content;
    showTemplateModal = false;
    dispatch('notification', { message: `已应用模板: ${template.name}`, type: 'success' });
  }

  function formatJson() {
    try {
      const parsed = JSON.parse(message.value);
      message.value = JSON.stringify(parsed, null, 2);
    } catch (error) {
      dispatch('notification', { message: 'JSON格式错误: ' + error.message, type: 'error' });
    }
  }

  function minifyJson() {
    try {
      const parsed = JSON.parse(message.value);
      message.value = JSON.stringify(parsed);
    } catch (error) {
      dispatch('notification', { message: 'JSON格式错误: ' + error.message, type: 'error' });
    }
  }

  function clearForm() {
    // 只清空消息内容，保留主题
    message.key = '';
    message.value = '';
    message.headers = {};
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

  <!-- Message Form -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" /></svg>
        <span>发送消息</span>
      </h2>
      <div class="form-control">
        <label for="producer-topic" class="label">
          <span class="label-text">主题 / 队列</span>
          {#if $selectedConnection && $selectedConnection.type === 'rocketmq'}
            <span class="label-text-alt text-info">RocketMQ: 请直接输入主题名称</span>
          {:else if availableTopics.length > 0}
            <div class="dropdown dropdown-end" class:dropdown-open={showTopicDropdown}>
              <div tabindex="0" role="button" class="btn btn-xs btn-outline"
                   on:click={() => showTopicDropdown = !showTopicDropdown}
                   on:keydown={(e) => e.key === 'Enter' && (showTopicDropdown = !showTopicDropdown)}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
                选择主题
              </div>
              <ul class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-64 max-h-60 overflow-y-auto border">
                {#each availableTopics as topic}
                  <li>
                    <button
                      class="text-left justify-start w-full"
                      class:opacity-60={isSystemTopic(topic)}
                      class:cursor-not-allowed={isSystemTopic(topic)}
                      class:hover:bg-base-200={!isSystemTopic(topic)}
                      disabled={isSystemTopic(topic)}
                      title={isSystemTopic(topic) ? '系统主题不可选择' : ''}
                      on:click={() => {
                        if (!isSystemTopic(topic)) {
                          message.topic = topic;
                          showTopicDropdown = false;
                        }
                      }}
                    >
                      <div class="flex items-center justify-between w-full">
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
                    </button>
                  </li>
                {/each}
              </ul>
            </div>
          {/if}
        </label>
        <input id="producer-topic" type="text" bind:value={message.topic} class="input input-bordered w-full" placeholder="输入主题名称或从上方选择" disabled={!isOnline} />
      </div>
      <div class="form-control">
        <label for="producer-value" class="label">
          <span class="label-text">消息内容</span>
          {#if messageTemplates.length > 0}
            <button class="btn btn-xs btn-outline" on:click={openTemplateModal}>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              使用模板
            </button>
          {/if}
        </label>
        <textarea id="producer-value" bind:value={message.value} class="textarea textarea-bordered h-48 font-mono" disabled={!isOnline}></textarea>
        <div class="label">
          <span class="label-text-alt"></span>
          <div class="flex space-x-2">
            <button class="btn btn-xs btn-ghost" on:click={formatJson} disabled={!message.value}>格式化JSON</button>
            <button class="btn btn-xs btn-ghost" on:click={minifyJson} disabled={!message.value}>压缩JSON</button>
            <button class="btn btn-xs btn-ghost" on:click={clearForm}>清空</button>
          </div>
        </div>
      </div>
      <div class="collapse collapse-arrow bg-base-200">
        <input type="checkbox" id="advanced-toggle" bind:checked={showAdvanced} /> 
        <label for="advanced-toggle" class="collapse-title text-md font-medium">高级选项</label>
        <div class="collapse-content">
          <div class="form-control">
            <label for="producer-key" class="label"><span class="label-text">消息 Key</span></label>
            <input id="producer-key" type="text" bind:value={message.key} class="input input-bordered w-full" disabled={!isOnline} />
          </div>
          <div class="form-control mt-4">
            <label class="label"><span class="label-text">消息头 (Headers)</span></label>
            <div class="space-y-2">
              {#each Object.entries(message.headers) as [key, value]}
                <div class="flex items-center space-x-2">
                  <input type="text" value={key} class="input input-sm input-bordered flex-1" disabled />
                  <input type="text" value={value} class="input input-sm input-bordered flex-1" disabled />
                  <button class="btn btn-sm btn-ghost btn-circle" on:click={() => removeHeader(key)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
                  </button>
                </div>
              {/each}
              <div class="flex items-center space-x-2">
                <input type="text" bind:value={headerKey} class="input input-sm input-bordered flex-1" placeholder="Key" id="header-key-input" />
                <input type="text" bind:value={headerValue} class="input input-sm input-bordered flex-1" placeholder="Value" id="header-value-input" />
                <button class="btn btn-sm btn-primary" on:click={addHeader}>添加</button>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="card-actions justify-end">
        <button class="btn btn-primary" on:click={sendMessage} disabled={!isOnline || sending}>
          {#if sending}<span class="loading loading-spinner"></span>{/if}
          发送
        </button>
      </div>
    </div>
  </div>
</div>

<!-- Template Selection Modal -->
{#if showTemplateModal}
  <div class="modal modal-open">
    <div class="modal-box w-11/12 max-w-4xl">
      <h3 class="font-bold text-lg mb-4">选择消息模板</h3>

      {#if messageTemplates.length === 0}
        <div class="text-center py-12">
          <div class="text-6xl mb-4">📝</div>
          <h4 class="text-lg font-semibold mb-2">暂无模板</h4>
          <p class="text-base-content/60">请先在模板管理页面创建消息模板。</p>
        </div>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-96 overflow-y-auto">
          {#each messageTemplates as template}
            <div class="card bg-base-200 hover:bg-base-300 cursor-pointer transition-colors"
                 on:click={() => useTemplate(template)}
                 on:keydown={(e) => e.key === 'Enter' && useTemplate(template)}
                 tabindex="0"
                 role="button"
                 aria-label="使用模板 {template.name}">
              <div class="card-body p-4">
                <h4 class="card-title text-base">{template.name}</h4>
                <div class="bg-base-100 rounded p-2 mt-2">
                  <pre class="text-xs text-base-content/80 whitespace-pre-wrap break-all max-h-20 overflow-y-auto font-mono">{template.content}</pre>
                </div>
                <div class="text-xs text-base-content/60 mt-2">
                  {new Date(template.createdAt).toLocaleDateString('zh-CN')}
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <div class="modal-action">
        <button class="btn btn-outline" on:click={closeTemplateModal}>取消</button>
      </div>
    </div>
  </div>
{/if}