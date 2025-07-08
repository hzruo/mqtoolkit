import { EventsOn } from '../wailsjs/runtime';
import { consumerMessages, consumerState } from './store.js';
import { get } from 'svelte/store';

// 全局事件监听器管理器
class EventManager {
  constructor() {
    this.listenersSetup = false;
    this.notificationCallback = null;
    this.stopConsumingCallback = null;
  }

  // 设置事件监听器
  setupEventListeners(notificationCallback, stopConsumingCallback) {
    if (this.listenersSetup) return;

    console.log('Setting up global event listeners');
    
    this.notificationCallback = notificationCallback;
    this.stopConsumingCallback = stopConsumingCallback;

    // 消息接收事件
    EventsOn('message:received', (message) => {
      console.log('Received message:', message);

      // 检查消息数量限制
      const currentMessages = get(consumerMessages);
      const currentState = get(consumerState);
      const maxMessages = currentState.maxMessages || 100;

      if (currentMessages.length >= maxMessages) {
        if (this.stopConsumingCallback) {
          this.stopConsumingCallback();
        }
        if (this.notificationCallback) {
          this.notificationCallback(`已达到最大消息数量限制 (${maxMessages})`, 'warning');
        }
        return;
      }

      consumerMessages.update(msgs => [message, ...msgs]);
    });

    // 消费错误事件
    EventsOn('consumer:error', (error) => {
      console.log('Received consumer error:', error);
      // 检查是否是正常的停止操作导致的错误
      const errorMsg = error.error || '';
      const isNormalShutdown = errorMsg.includes('connection closed') ||
                              errorMsg.includes('use of closed network connection') ||
                              errorMsg.includes('CONN_');

      if (!isNormalShutdown && this.notificationCallback) {
        this.notificationCallback(`消费错误: ${error.error}`, 'error');
      }
      
      if (this.stopConsumingCallback) {
        this.stopConsumingCallback();
      }
    });

    this.listenersSetup = true;
  }

  // 更新回调函数
  updateCallbacks(notificationCallback, stopConsumingCallback) {
    this.notificationCallback = notificationCallback;
    this.stopConsumingCallback = stopConsumingCallback;
  }
}

// 导出单例实例
export const eventManager = new EventManager();
