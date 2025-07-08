<script>
  import { onMount } from 'svelte';
  import { selectedConnection, testResults } from './store.js';
  import ConnectionManager from './components/ConnectionManager.svelte';
  import MessageProducer from './components/MessageProducer.svelte';
  import MessageConsumer from './components/MessageConsumer.svelte';
  import HistoryViewer from './components/HistoryViewer.svelte';
  import LogViewer from './components/LogViewer.svelte';
  import TopicManager from './components/TopicManager.svelte';
  import TemplateManager from './components/TemplateManager.svelte';
  import About from './components/About.svelte';
  import { eventManager } from './eventManager.js';

  let activeTab = 'connections';
  let notification = null;
  let isSelectedConnectionOnline = false;
  let currentTheme = 'light';

  // Subscribe to store changes to compute derived state
  $: {
    if ($selectedConnection && $testResults && $testResults[$selectedConnection.id]) {
      isSelectedConnectionOnline = $testResults[$selectedConnection.id].success;
    } else {
      isSelectedConnectionOnline = false;
    }
  }

  const tabs = {
    connections: { label: '连接管理', component: ConnectionManager },
    producer: { label: '消息发送', component: MessageProducer },
    consumer: { label: '消息消费', component: MessageConsumer },
    topics: { label: '主题/队列', component: TopicManager },
    templates: { label: '消息模板', component: TemplateManager },
    history: { label: '历史记录', component: HistoryViewer },
    logs: { label: '日志查看', component: LogViewer },
    about: { label: '关于', component: About },
  };

  const icons = {
    connections: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
    producer: 'M12 19l9 2-9-18-9 18 9-2zm0 0v-8',
    consumer: 'M7 16l-4-4m0 0l4-4m-4 4h18',
    topics: 'M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m-1 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3h2m-4 17h4m-7-7h2m-4 4h2m4-4h2m4 4h2m-4-4h2m-4-4h2',
    templates: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10',
    history: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    logs: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
    about: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  };

  function showNotification(message, type = 'info') {
    notification = { message, type };
    setTimeout(() => {
      notification = null;
    }, 3000);
  }

  function handleNotification(event) {
    showNotification(event.detail.message, event.detail.type);
  }

  // 主题相关功能
  const themes = [
    { value: 'light', label: '浅色主题' },
    { value: 'dark', label: '深色主题' },
    { value: 'cupcake', label: '杯子蛋糕' },
    { value: 'garden', label: '花园' },
    { value: 'forest', label: '森林' },
    { value: 'synthwave', label: '合成波' },
    { value: 'valentine', label: '情人节' },
    { value: 'aqua', label: '水蓝色' },
    { value: 'retro', label: '复古' },
    { value: 'cyberpunk', label: '赛博朋克' },
  ];

  let showThemeDropdown = false;

  function changeTheme(theme) {
    currentTheme = theme;
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('theme', theme);
    // 选择主题后关闭下拉菜单
    showThemeDropdown = false;
  }

  // 初始化全局事件管理器
  onMount(() => {
    console.log('App mounted, initializing event manager');

    // 恢复保存的主题
    const savedTheme = localStorage.getItem('theme') || 'light';
    changeTheme(savedTheme);

    eventManager.setupEventListeners(showNotification, () => {
      // 这里可以添加全局的停止消费逻辑
      console.log('Global stop consuming called');
    });
  });
</script>

<div class="flex h-screen bg-base-200 font-sans">
  <!-- Sidebar -->
  <aside class="w-60 bg-base-100 text-base-content flex-shrink-0 flex flex-col">
    <div class="flex items-center justify-center h-16 bg-base-300 shadow-md">
      <div class="avatar placeholder mr-2">
        <div class="bg-primary text-primary-content rounded-full w-8">
          <span class="text-sm font-bold">MQ</span>
        </div>
      </div>
      <h1 class="text-xl font-bold">MQ Toolkit</h1>
    </div>

    <ul class="menu p-4 flex-1">
      {#each Object.entries(tabs) as [key, tab] (key)}
        <li>
          <button on:click={() => activeTab = key} class:active={activeTab === key} class="w-full justify-start">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={icons[key]}></path>
            </svg>
            {tab.label}
          </button>
        </li>
      {/each}
    </ul>

    <div class="p-4">
      <div class="dropdown dropdown-top w-full" class:dropdown-open={showThemeDropdown}>
        <div
          tabindex="0"
          role="button"
          class="btn btn-outline w-full"
          on:click={() => showThemeDropdown = !showThemeDropdown}
          on:keydown={(e) => e.key === 'Enter' && (showThemeDropdown = !showThemeDropdown)}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zM21 5a2 2 0 00-2-2h-4a2 2 0 00-2 2v12a4 4 0 004 4h4a2 2 0 002-2V5z" />
          </svg>
          主题切换
        </div>
        {#if showThemeDropdown}
          <ul class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-full max-h-60 overflow-y-auto">
            {#each themes as theme}
              <li>
                <button
                  class:bg-primary={currentTheme === theme.value}
                  class:text-primary-content={currentTheme === theme.value}
                  on:click={() => changeTheme(theme.value)}
                >
                  {theme.label}
                  {#if currentTheme === theme.value}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  {/if}
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    </div>
  </aside>

  <!-- Main Content -->
  <main class="flex-1 flex flex-col overflow-hidden">
    <header class="bg-base-100 p-4 shadow-md z-10">
      <h2 class="text-2xl font-semibold">{tabs[activeTab].label}</h2>
    </header>

    <div class="flex-1 overflow-y-auto p-6">
      <svelte:component 
        this={tabs[activeTab].component}
        isOnline={isSelectedConnectionOnline}
        on:notification={handleNotification}
      />
    </div>
  </main>

  <!-- Notification System -->
  {#if notification}
    <div class="toast toast-top toast-center z-50">
      <div class="alert alert-{notification.type} shadow-lg">
        <div>
          <span>{notification.message}</span>
        </div>
      </div>
    </div>
  {/if}
</div>