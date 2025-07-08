<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ListTemplates, CreateTemplate, UpdateTemplate, DeleteTemplate } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let templates = [];
  let loading = false;
  let showForm = false;
  let editingTemplate = null;
  let form = { id: '', name: '', content: '' };
  let showDeleteConfirm = false;
  let templateToDelete = null;

  async function loadTemplates() {
    try {
      loading = true;
      templates = await ListTemplates() || [];
    } catch (error) {
      dispatch('notification', { message: '加载模板失败: ' + error, type: 'error' });
    } finally {
      loading = false;
    }
  }

  function openForm(template = null) {
    if (template) {
      editingTemplate = template;
      form = { ...template };
    } else {
      editingTemplate = null;
      form = { id: `new_${Date.now()}`, name: '', content: '' };
    }
    showForm = true;
  }

  function closeForm() {
    showForm = false;
  }

  async function handleSubmit() {
    if (!form.name || !form.content) {
      dispatch('notification', { message: '请填写模板名称和内容', type: 'error' });
      return;
    }

    try {
      const templateData = {...form};
      if(!editingTemplate) delete templateData.id;

      if (editingTemplate) {
        await UpdateTemplate(templateData.id, templateData.name, templateData.content);
        dispatch('notification', { message: '模板更新成功', type: 'success' });
      } else {
        await CreateTemplate(templateData.name, templateData.content);
        dispatch('notification', { message: '模板创建成功', type: 'success' });
      }
      closeForm();
      loadTemplates();
    } catch (error) {
      dispatch('notification', { message: '保存模板失败: ' + error, type: 'error' });
    }
  }

  function startDeleteTemplate(template) {
    templateToDelete = template;
    showDeleteConfirm = true;
  }

  function cancelDelete() {
    templateToDelete = null;
    showDeleteConfirm = false;
  }

  async function confirmDelete() {
    if (!templateToDelete) return;

    try {
      await DeleteTemplate(templateToDelete.id);
      dispatch('notification', { message: '模板删除成功', type: 'success' });
      loadTemplates();
    } catch (error) {
      dispatch('notification', { message: '删除模板失败: ' + error, type: 'error' });
    } finally {
      cancelDelete();
    }
  }

  onMount(loadTemplates);
</script>

<div class="space-y-6">
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="flex justify-between items-center mb-4">
        <h2 class="card-title">消息模板管理</h2>
        <button class="btn btn-primary btn-sm" on:click={() => openForm()}>新建模板</button>
      </div>

      {#if loading}
        <div class="text-center py-8"><span class="loading loading-spinner"></span></div>
      {:else if templates.length === 0}
        <div class="text-center py-8">
          <p>暂无模板，点击“新建模板”来创建一个。</p>
        </div>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {#each templates as template (template.id)}
            <div class="card bg-base-200 shadow-lg hover:shadow-xl transition-shadow duration-200">
              <div class="card-body p-6">
                <div class="flex items-start justify-between mb-3">
                  <h3 class="card-title text-lg font-semibold text-primary">{template.name}</h3>
                  <div class="badge badge-outline badge-sm">模板</div>
                </div>

                <div class="mb-4">
                  <div class="bg-base-100 rounded-lg p-3 border">
                    <pre class="text-xs text-base-content/80 whitespace-pre-wrap break-all max-h-32 overflow-y-auto font-mono">{template.content}</pre>
                  </div>
                </div>

                <div class="flex items-center justify-between text-xs text-base-content/60 mb-4">
                  <span>创建时间</span>
                  <span>{new Date(template.createdAt).toLocaleDateString('zh-CN')}</span>
                </div>

                <div class="card-actions justify-end space-x-2">
                  <button class="btn btn-sm btn-outline btn-primary" on:click={() => openForm(template)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.536L16.732 3.732z" />
                    </svg>
                    编辑
                  </button>
                  <button class="btn btn-sm btn-outline btn-error" on:click={() => startDeleteTemplate(template)}>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    删除
                  </button>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>

<!-- Template Form Modal -->
{#if showForm}
  <div class="modal modal-open">
    <div class="modal-box w-11/12 max-w-2xl">
      <h3 class="font-bold text-lg mb-4">{editingTemplate ? '编辑' : '新建'}模板</h3>
      <div class="form-control">
        <label for="template-name-{form.id}" class="label"><span class="label-text">模板名称</span></label>
        <input id="template-name-{form.id}" type="text" bind:value={form.name} class="input input-bordered" />
      </div>
      <div class="form-control mt-4">
        <label for="template-content-{form.id}" class="label"><span class="label-text">模板内容</span></label>
        <textarea id="template-content-{form.id}" bind:value={form.content} class="textarea textarea-bordered h-48 font-mono"></textarea>
      </div>
      <div class="modal-action">
        <button class="btn btn-primary" on:click={handleSubmit}>保存</button>
        <button class="btn btn-outline" on:click={closeForm}>取消</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteConfirm && templateToDelete}
  <div class="modal modal-open">
    <div class="modal-box">
      <h3 class="font-bold text-lg">确认删除</h3>
      <p class="py-4">确定要删除模板 <span class="font-mono bg-base-200 px-2 py-1 rounded">"{templateToDelete.name}"</span> 吗？</p>
      <p class="text-warning text-sm">⚠️ 此操作不可恢复。</p>
      <div class="modal-action">
        <button class="btn btn-error" on:click={confirmDelete}>确认删除</button>
        <button class="btn btn-outline" on:click={cancelDelete}>取消</button>
      </div>
    </div>
  </div>
{/if}
