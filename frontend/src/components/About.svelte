<script>
  import { onMount } from 'svelte';
  import { BrowserOpenURL } from '../../wailsjs/runtime';
  import logoSvg from '../assets/logo.svg';

  let appVersion = '1.0.0';
  let buildTime = '';
  let goVersion = '';
  let wailsVersion = '';

  onMount(() => {
    // 这里可以从后端获取版本信息
    buildTime = new Date().toLocaleString('zh-CN');
    goVersion = 'Go 1.23+';
    wailsVersion = 'Wails v2.5+';
  });

  function openGitHub() {
    BrowserOpenURL('https://github.com/hzruo/mqtoolkit');
  }

  function openEmail() {
    BrowserOpenURL('mailto:hzruo@outlook.com');
  }
</script>

<div class="container mx-auto p-6">
  <div class="card bg-base-100 shadow-xl">
    <div class="card-body">
      <div class="text-center mb-8">
        <img src={logoSvg} alt="MQ Toolkit Logo" class="h-24 w-24 mx-auto mb-4 drop-shadow-lg" />
        <h2 class="text-3xl font-bold text-primary">关于 MQ Toolkit</h2>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- 应用信息 -->
        <div class="space-y-4">
          <h3 class="text-lg font-semibold">应用信息</h3>
          <div class="space-y-2">
            <div class="flex justify-between">
              <span class="text-base-content/70">版本:</span>
              <span class="font-mono">{appVersion}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/70">作者:</span>
              <span class="font-mono text-sm">hzruo</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/70">Go 版本:</span>
              <span class="font-mono">{goVersion}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-base-content/70">Wails 版本:</span>
              <span class="font-mono">{wailsVersion}</span>
            </div>
          </div>
        </div>

        <!-- 功能特性 -->
        <div class="space-y-4">
          <h3 class="text-lg font-semibold">功能特性</h3>
          <ul class="space-y-2">
            <li class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              支持 Kafka、RabbitMQ、RocketMQ
            </li>
            <li class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              消息生产与消费
            </li>
            <li class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              主题/队列管理
            </li>
            <li class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              消息模板管理
            </li>
            <li class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              操作历史记录
            </li>
            <li class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              实时日志查看
            </li>
          </ul>
        </div>
      </div>

      <!-- 技术栈 -->
      <div class="mt-6">
        <h3 class="text-lg font-semibold mb-4">技术栈</h3>
        <div class="flex flex-wrap gap-2">
          <div class="badge badge-primary">Go</div>
          <div class="badge badge-secondary">Wails</div>
          <div class="badge badge-accent">Svelte</div>
          <div class="badge badge-neutral">DaisyUI</div>
          <div class="badge badge-info">TailwindCSS</div>
          <div class="badge badge-success">SQLite</div>
        </div>
      </div>

      <!-- 开源信息 -->
      <div class="mt-6 p-4 bg-base-200 rounded-lg">
        <h3 class="text-lg font-semibold mb-2">开源许可</h3>
        <p class="text-base-content/70 text-sm">
          本项目基于 MIT 许可证开源，欢迎贡献代码和反馈问题。
        </p>
        <div class="mt-3 flex gap-2">
          <button class="btn btn-sm btn-outline" on:click={openGitHub}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
            </svg>
            GitHub
          </button>
          <button class="btn btn-sm btn-outline" on:click={openEmail}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            反馈
          </button>
        </div>
      </div>
    </div>
  </div>
</div>
