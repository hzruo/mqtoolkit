import { writable } from 'svelte/store';

/**
 * @typedef {import('../../wailsjs/go/models').types.ConnectionConfig} ConnectionConfig
 * @typedef {import('../../wailsjs/go/models').types.TestResult} TestResult
 */

/** @type {import('svelte/store').Writable<ConnectionConfig[]>} */
export const connections = writable([]);

/** @type {import('svelte/store').Writable<ConnectionConfig | null>} */
export const selectedConnection = writable(null);

/** @type {import('svelte/store').Writable<{[key: string]: TestResult}>} */
export const testResults = writable({});

export const loading = writable(false);

// 持久化主题选择
function createPersistedStore(key, defaultValue) {
  const stored = localStorage.getItem(key);
  const initial = stored ? JSON.parse(stored) : defaultValue;

  const store = writable(initial);

  store.subscribe(value => {
    localStorage.setItem(key, JSON.stringify(value));
  });

  return store;
}

// 生产者选择的主题
export const selectedProducerTopic = createPersistedStore('selectedProducerTopic', '');

// 消费者选择的主题
export const selectedConsumerTopics = createPersistedStore('selectedConsumerTopics', '');

// 消费状态
export const consumerState = createPersistedStore('consumerState', {
  consuming: false,
  subscriptionId: null,
  groupId: 'mq-toolkit-consumer',
  fromBeginning: false,
  maxMessages: 100
});

// 消费消息列表（持久化）
export const consumerMessages = createPersistedStore('consumerMessages', []);
