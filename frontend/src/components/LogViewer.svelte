<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime';
  import { GetLogs, SaveFile } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let logs = [];
  let loading = false;

  let logLevel = 'all';
  let searchTerm = '';
  let maxLogs = 500;
  let following = true;
  let logContainer;

  // 加载初始日志
  async function loadInitialLogs() {
    try {
      loading = true;
      const initialLogs = await GetLogs();
      logs = initialLogs || [];
      scrollToBottom();
    } catch (error) {
      dispatch('notification', { message: '加载初始日志失败: ' + error, type: 'error' });
    } finally {
      loading = false;
    }
  }

  // 在模块级别设置事件监听器
  EventsOn('log:new', (entry) => {
    logs = [...logs, entry].slice(-maxLogs);
    if (following) {
      scrollToBottom();
    }
  });

  onMount(() => {
    loadInitialLogs();

    return () => {
      EventsOff('log:new');
    };
  });

  // 过滤日志
  $: filteredLogs = logs.filter(log => {
    const matchesLevel = logLevel === 'all' || log.level.toLowerCase() === logLevel;
    const matchesSearch = !searchTerm || 
      (log.message && log.message.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (log.source && log.source.toLowerCase().includes(searchTerm.toLowerCase()));
    
    return matchesLevel && matchesSearch;
  });

  // 格式化时间
  function formatTime(timestamp) {
    return new Date(timestamp).toLocaleString('zh-CN', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3
    });
  }

  // 获取日志级别样式
  function getLevelClass(level) {
    const classes = {
      'DEBUG': 'text-base-content/60',
      'INFO': 'text-info',
      'WARN': 'text-warning',
      'ERROR': 'text-error',
      'FATAL': 'text-error font-bold'
    };
    return classes[level] || 'text-base-content';
  }

  // 获取日志级别图标
  function getLevelIcon(level) {
    const icons = {
      'DEBUG': '🔍',
      'INFO': 'ℹ️',
      'WARN': '⚠️',
      'ERROR': '❌',
      'FATAL': '💀'
    };
    return icons[level] || '📝';
  }

  // 清空日志
  function clearLogs() {
    logs = [];
    dispatch('notification', { message: '日志已清空', type: 'success' });
  }

  // 导出日志
  async function exportLogs() {
    try {
      const data = filteredLogs.map(log =>
        `[${formatTime(log.timestamp)}] [${log.level}] [${log.source}] ${log.message}`
      ).join('\n');

      const filename = `logs-${new Date().toISOString().split('T')[0]}.txt`;

      const savedPath = await SaveFile(filename, data);

      dispatch('notification', {
        message: `日志已导出 (${filteredLogs.length} 条日志)\n保存位置: ${savedPath}`,
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

  // 滚动到底部
  function scrollToBottom() {
    setTimeout(() => {
      if (logContainer) {
        logContainer.scrollTop = logContainer.scrollHeight;
      }
    }, 50);
  }

  // 切换跟随模式
  function toggleFollowing() {
    following = !following;
    if (following) {
      scrollToBottom();
    }
  }

  // 复制日志
  function copyLog(log) {
    const text = `[${formatTime(log.timestamp)}] [${log.level}] [${log.source}] ${log.message}`;
    navigator.clipboard.writeText(text).then(() => {
      dispatch('notification', { message: '日志已复制到剪贴板', type: 'success' });
    });
  }
</script>

<!-- ... (script) ... -->
<div class="space-y-6">
  <!-- Controls -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center">
        <h2 class="card-title">系统日志</h2>
        <div class="flex items-center space-x-2">
          <input type="text" bind:value={searchTerm} placeholder="搜索..." class="input input-bordered input-sm" />
          <select bind:value={logLevel} class="select select-sm select-bordered">
            <option value="all">全部</option>
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warn">Warn</option>
            <option value="error">Error</option>
          </select>
        </div>
      </div>
    </div>
  </div>

  <!-- Log View -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body p-0">
      <div class="flex justify-between items-center p-4 border-b">
        <div class="flex items-center space-x-2">
          <span class="text-sm text-base-content/60">共 {filteredLogs.length} 条日志</span>
          {#if following}
            <div class="badge badge-success badge-sm whitespace-nowrap">跟随模式</div>
          {/if}
        </div>
        <div class="flex space-x-2">
          <button class="btn btn-xs btn-outline" on:click={toggleFollowing}>
            {following ? '停止跟随' : '跟随日志'}
          </button>
          <button class="btn btn-xs btn-outline" on:click={exportLogs} disabled={filteredLogs.length === 0}>导出</button>
          <button class="btn btn-xs btn-error btn-outline" on:click={clearLogs} disabled={logs.length === 0}>清空</button>
        </div>
      </div>

      <div bind:this={logContainer} class="h-96 overflow-y-auto font-mono text-sm p-4 space-y-1">
        {#if filteredLogs.length === 0}
          <div class="text-center py-12">
            <div class="text-6xl mb-4">📋</div>
            <h3 class="text-lg font-semibold mb-2">暂无日志</h3>
            <p class="text-base-content/60">应用程序运行时的日志将显示在这里。</p>
          </div>
        {:else}
          {#each filteredLogs as log (log.timestamp)}
            <div class="flex items-start space-x-2 hover:bg-base-200 p-1 rounded group">
              <span class="text-xs text-base-content/40 w-20 flex-shrink-0">
                {formatTime(log.timestamp)}
              </span>
              <span class="w-12 flex-shrink-0 {getLevelClass(log.level)}">
                {getLevelIcon(log.level)} {log.level}
              </span>
              <span class="text-xs text-base-content/60 w-20 flex-shrink-0">
                [{log.source}]
              </span>
              <span class="flex-1 break-all">
                {log.message}
              </span>
              <button
                class="btn btn-xs btn-ghost opacity-0 group-hover:opacity-100 transition-opacity"
                on:click={() => copyLog(log)}
                title="复制日志"
              >
                📋
              </button>
            </div>
          {/each}
        {/if}
      </div>
    </div>
  </div>
</div>
