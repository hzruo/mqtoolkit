<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { GetHistory, ClearHistory, SaveFile } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let history = [];
  let loading = false;
  let currentPage = 1;
  let pageSize = 20;
  let totalRecords = 0;
  let searchTerm = '';
  let filterType = 'all';
  let selectedRecord = null;
  let showClearConfirm = false;

  // 加载历史记录
  async function loadHistory() {
    try {
      loading = true;
      const offset = (currentPage - 1) * pageSize;
      const records = await GetHistory(pageSize, offset);
      history = records || [];
      // This should be improved to get the total count from the backend
      totalRecords = history.length + offset; 
    } catch (error) {
      dispatch('notification', { message: '加载历史记录失败: ' + error, type: 'error' });
    } finally {
      loading = false;
    }
  }

  // 过滤历史记录
  $: filteredHistory = history.filter(record => {
    const matchesSearch = !searchTerm ||
      record.type.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (record.topic && record.topic.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (record.message && record.message.toLowerCase().includes(searchTerm.toLowerCase()));

    const matchesType = filterType === 'all' || record.type === filterType;

    return matchesSearch && matchesType;
  });

  // 格式化时间
  function formatTime(timestamp) {
    return new Date(timestamp).toLocaleString('zh-CN');
  }

  // 获取操作类型图标
  function getOperationIcon(operation) {
    const icons = {
      'produce': '📤',
      'consume': '📥',
      'test_connection': '🔍',
      'create_connection': '🔗',
      'delete_connection': '🗑️'
    };
    return icons[operation] || '📋';
  }

  // 获取状态样式
  function getStatusClass(status) {
    return status ? 'badge-success' : 'badge-error';
  }

  // 查看详情
  function viewDetails(record) {
    selectedRecord = record;
  }

  // 关闭详情
  function closeDetails() {
    selectedRecord = null;
  }

  // 导出历史记录
  async function exportHistory() {
    try {
      const data = JSON.stringify(filteredHistory, null, 2);
      const filename = `history-${new Date().toISOString().split('T')[0]}.json`;

      const savedPath = await SaveFile(filename, data);

      dispatch('notification', {
        message: `历史记录已导出 (${filteredHistory.length} 条记录)\n保存位置: ${savedPath}`,
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

  function startClearHistory() {
    showClearConfirm = true;
  }

  function cancelClear() {
    showClearConfirm = false;
  }

  // 清空历史记录
  async function confirmClearHistory() {
    try {
      await ClearHistory();
      dispatch('notification', { message: '历史记录已清空', type: 'success' });
      loadHistory();
    } catch (error) {
      dispatch('notification', { message: '清空历史记录失败: ' + error, type: 'error' });
    } finally {
      showClearConfirm = false;
    }
  }

  // 刷新
  function refresh() {
    loadHistory();
  }

  // 分页
  function goToPage(page) {
    currentPage = page;
    loadHistory();
  }

  onMount(() => {
    loadHistory();
  });
</script>

<!-- ... (script) ... -->
<div class="space-y-6">
  <!-- Controls -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center">
        <h2 class="card-title">历史记录</h2>
        <div class="flex items-center space-x-2">
          <select bind:value={filterType} class="select select-bordered select-sm">
            <option value="all">全部类型</option>
            <option value="produce">发送消息</option>
            <option value="consume">消费消息</option>
            <option value="test_connection">连接测试</option>
          </select>
          <input type="text" bind:value={searchTerm} placeholder="搜索..." class="input input-bordered input-sm" />
          <button class="btn btn-sm btn-error btn-outline" on:click={startClearHistory} disabled={history.length === 0}>清空历史</button>
        </div>
      </div>
    </div>
  </div>

  <!-- History Table -->
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      {#if loading}
        <div class="text-center py-12">
          <span class="loading loading-spinner loading-lg"></span>
          <p class="mt-4 text-base-content/60">正在加载历史记录...</p>
        </div>
      {:else if filteredHistory.length === 0}
        <div class="text-center py-12">
          <div class="text-6xl mb-4">📋</div>
          <h3 class="text-lg font-semibold mb-2">暂无历史记录</h3>
          <p class="text-base-content/60">开始使用应用程序后，操作历史将显示在这里。</p>
        </div>
      {:else}
        <div class="overflow-x-auto">
          <table class="table w-full">
            <thead>
              <tr>
                <th>时间</th>
                <th>类型</th>
                <th>主题</th>
                <th>状态</th>
                <th>延迟</th>
                <th>消息</th>
              </tr>
            </thead>
            <tbody>
              {#each filteredHistory as record (record.id)}
                <tr class="hover">
                  <td class="text-sm">{formatTime(record.created)}</td>
                  <td>
                    <div class="badge badge-outline whitespace-nowrap">
                      <span class="mr-1">{getOperationIcon(record.type)}</span>
                      <span>{record.type}</span>
                    </div>
                  </td>
                  <td class="font-mono text-sm">{record.topic || '-'}</td>
                  <td>
                    <div class="badge {getStatusClass(record.success)} whitespace-nowrap">
                      {record.success ? '成功' : '失败'}
                    </div>
                  </td>
                  <td class="text-sm">{record.latency}ms</td>
                  <td class="text-sm max-w-xs truncate" title={record.message}>
                    {record.message}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex justify-between items-center mt-4">
          <div class="text-sm text-base-content/60">
            显示 {filteredHistory.length} 条记录
          </div>
          <div class="flex space-x-2">
            <button class="btn btn-sm btn-outline" on:click={refresh}>刷新</button>
            <button class="btn btn-sm btn-outline" on:click={exportHistory} disabled={filteredHistory.length === 0}>导出</button>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<!-- Clear Confirmation Modal -->
{#if showClearConfirm}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg">确认清空</h3>
      <p class="py-4">确定要清空所有历史记录吗？</p>
      <p class="text-warning text-sm">⚠️ 此操作不可恢复，所有历史记录都将被永久删除。</p>
      <div class="modal-action">
        <button class="btn btn-error" on:click={confirmClearHistory}>确认清空</button>
        <button class="btn btn-outline" on:click={cancelClear}>取消</button>
      </div>
    </div>
  </div>
{/if}
