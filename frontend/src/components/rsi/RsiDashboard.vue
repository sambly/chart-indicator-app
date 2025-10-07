<script setup lang="ts">
import { reactive, onMounted,ref,watchEffect   } from 'vue'
import { NButton, NInput, NCard, NDatePicker, NSpin,NSelect,NInputNumber } from 'naive-ui'
import { buildSniper } from './buildSniper'


const chartEl = ref<HTMLElement | null>(null)

interface RSIData {
  symbol: string
  startDate: string | null
  endDate: string | null
  interval: string
  config: RsiConfig
  currentOpti: OptimizationResult
  optimization: OptimizationResult
  chartData?: any
  signalBuyPoints?: any
  signalSellPoints?: any

}

interface RsiConfig {
  rsiLength: number
  emaSlowLength: number
  rsiBuyLevel: number
  rsiExitLevel: number
  minBarsBetweenTrades: number
}

interface OptimizationResult {
  profit: number
  trades: number
  winRate: number
  drawdown: number
  winRatePercent: number
  countSignalBuy: number
  countSignalSell: number
}

const state = reactive<RSIData>({
  symbol: '',
  startDate: null,
  endDate: null,
  interval: '',
  config: {
    rsiLength: 0,
    emaSlowLength: 0,
    rsiBuyLevel: 0,
    rsiExitLevel: 0,
    minBarsBetweenTrades: 0
  },
  currentOpti: {
    profit: 0,
    trades: 0,
    winRate: 0,
    drawdown: 0,
    winRatePercent: 0,
    countSignalBuy: 0,
    countSignalSell: 0
  },
    optimization: {
    profit: 0,
    trades: 0,
    winRate: 0,
    drawdown: 0,
    winRatePercent: 0,
    countSignalBuy: 0,
    countSignalSell: 0
  },
  chartData: null,
  signalBuyPoints: null,
  signalSellPoints: null
})


const intervalOptions = [
  { label: '1 минута', value: '60' },
  { label: '5 минут', value: '300' },
  { label: '15 минут', value: '900' },
  { label: '1 час', value: '3600' },
  { label: 'День', value: 'd' }
]


let chartInstance: any = null  // хранит экземпляр графика

watchEffect(() => {
  if (state.chartData && chartEl.value) {
    if (!chartInstance) {
      chartInstance = buildSniper(
        state.chartData,
        state.signalBuyPoints || [],
        state.signalSellPoints || [],
        chartEl.value
      )
    } else {
      chartInstance.update(
        state.chartData,
        state.signalBuyPoints || [],
        state.signalSellPoints || []
      )
    }
  }
})


const isLoading = reactive({ value: false })


// ✅ 1. Загрузка дефолтных данных
async function fetchData() {
  isLoading.value = true
  try {
    const res = await fetch('/api/rsi/default-data')
    if (!res.ok) throw new Error('Ошибка загрузки данных')
    const data: RSIData = await res.json()

    // Обновляем реактивное состояние
    state.symbol = data.symbol
    state.startDate = data.startDate
    state.endDate = data.endDate
    state.interval = data.interval
    Object.assign(state.config, data.config ?? {})
    Object.assign(state.currentOpti, data.currentOpti ?? {})
    Object.assign(state.optimization, data.optimization ?? {})
    state.chartData = data.chartData
    state.signalBuyPoints = data.signalBuyPoints
    state.signalSellPoints = data.signalSellPoints

    console.log('fetchData: Данные получены:', state)

  } catch (err) {
    console.error(err)
  } finally {
    isLoading.value = false
  }
}

// ✅ 2. Обновление котировок и пересчёт RSI
async function applyMain() {
  isLoading.value = true
  try {
    const res = await fetch('/api/rsi/update', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        symbol: state.symbol,
        start_date: state.startDate,
        end_date: state.endDate,
        interval: state.interval,
      }),
    })

    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Неизвестная ошибка сервера')
    }

    state.chartData = data.chartData
    state.signalBuyPoints = data.signalBuyPoints
    state.signalSellPoints = data.signalSellPoints

    console.log('applyMain: Данные получены:', data)

  } catch (err) {
    console.error('applyMain error:', err)
  } finally {
    isLoading.value = false
  }
}


// ✅ 3. Применить конфиг RSI к текущим данным
async function applyConfig() {
  isLoading.value = true
  try {
    const res = await fetch('/api/rsi/apply-config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(state.config),
    })

    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Неизвестная ошибка сервера')
    }

    state.chartData = data.chartData
    state.signalBuyPoints = data.signalBuyPoints
    state.signalSellPoints = data.signalSellPoints

    console.log('applyConfig: Данные получены:', data)

  } catch (err) {
    console.error('applyConfig error:', err)
  } finally {
    isLoading.value = false
  }
}

// ✅ 4. Загрузить дефолтную конфигурацию
async function loadDefaultConfig() {
  isLoading.value = true
  try {
    const res = await fetch('/api/rsi/default-config')
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Неизвестная ошибка сервера')
    }
    Object.assign(state.config, data.config)
    console.log('loadDefaultConfig: Данные получены:', data)
  } catch (err) {
    console.error('loadDefaultConfig error:', err)
  } finally {
    isLoading.value = false
  }
}


// ✅ 5. Сохранить конфигурацию в файл
async function saveConfig() {
  isLoading.value = true
  try {
    const res = await fetch('/api/rsi/save-config', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(state.config),
    })   
   
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Неизвестная ошибка сервера')
    }
   
    console.log('saveConfig: Данные сохранены')

  } catch (err) {
    console.error('saveConfig error:', err)
  } finally {
    isLoading.value = false
  }
}

// ✅ 6. Выполнить текущую оптимизацию
async function evaluateCurrent() {
  try {
    const res = await fetch('/api/rsi/evaluate', { method: 'POST' })
    if (!res.ok) throw new Error('Ошибка оценки стратегии')
    const data = await res.json()
    Object.assign(state.currentOpti, data.currentOpti)
    console.log('evaluateCurrent: Произведенна оптимизация')
  } catch (err) {
    console.error('evaluateCurrent error:', err)
  }
}

// ✅ 7. Выполнить полную оптимизацию
async function optimizeRSI() {
  isLoading.value = true
  try {
    const res = await fetch('/api/rsi/optimize', { method: 'POST' })
    if (!res.ok) throw new Error('Ошибка оптимизации стратегии')
    const data = await res.json()
    Object.assign(state.optimization, data.optimization)
    Object.assign(state.config, data.config)
    state.chartData = data.chartData
    state.signalBuyPoints = data.signalBuyPoints
    state.signalSellPoints = data.signalSellPoints
    console.log('optimizeRSI: Произведенна полная оптимизация', data)
  } catch (err) {
    console.error('optimizeRSI error:', err)
  } finally {
    isLoading.value = false
  }
}

onMounted(fetchData)

</script>

<template>

  <div style="display: flex; flex-direction: column; align-items: center; min-height: 100vh; padding: 16px;">
    <NCard 
      title="RSI + EMA Dashboard" 
      size="large" 
      style="width: 1104px; max-width: 100%; margin-bottom: 16px;"
    >
      
      <!-- Лоадер -->
      <div v-if="isLoading.value" style="text-align:center; padding:20px;">
        <NSpin size="large">
          <template #description>
            Загрузка данных...
          </template>
        </NSpin>
      </div>

      <!-- Контент -->
      <div v-else style="display:flex; gap:16px; justify-content: center;">

        <!-- Основные параметры -->
        <NCard title="Основные параметры" size="small" style="width:250px; height:380px;">
          <div style="height:200px; overflow-y:auto; display:flex; flex-direction:column; gap:8px; padding-right:4px;">
              <NInput v-model:value="state.symbol" placeholder="Тикер"/>
              <NDatePicker 
              v-model:formatted-value="state.startDate" 
              type="date" 
              placeholder="Начальная дата"
              value-format="yyyy-MM-dd"
              />
              <NDatePicker 
              v-model:formatted-value="state.endDate" 
              type="date" 
              placeholder="Конечная дата"
              value-format="yyyy-MM-dd"
              />
              <NSelect
                  v-model:value="state.interval"
                  :options="intervalOptions"
                  placeholder="Интервал"
              />
          </div>
          <NButton @click="applyMain" type="primary" style="width:100%;">Обновить</NButton>
        </NCard>


        <!-- Настройки индикатора -->
        <NCard title="Настройки индикатора" size="small" style="width:250px; height:380px;">
          <div style="height:200px; overflow-y:auto; display:flex; flex-direction:column; gap:1px; padding-right:4px;">
              <div
                v-for="key in Object.keys(state.config)"
                :key="key"
                style="margin-bottom:8px;"
              >
                <label style="display:block; margin-bottom:4px; font-size:12px; color:#666;">
                  {{ key }}
                </label>
                <NInputNumber v-model:value="state.config[key as keyof typeof state.config]" :min="0" style="width: 100%;" />
              </div>
          </div>
            <div style="display:flex; flex-direction:column; gap:2px; margin-top:12px;">
              <NButton @click="loadDefaultConfig" type="primary" style="width:100%;">Default</NButton>
              <NButton @click="applyConfig" type="primary" style="width:100%;">Применить</NButton>
              <NButton @click="saveConfig" type="primary" style="width:100%;">Запись в файл</NButton> 
            </div>    
        </NCard>


         <!-- Результаты -->
        <NCard title="Результаты" size="small" style="width:250px; height:380px;">
          <div style="height:200px; overflow-y:auto; display:flex; flex-direction:column; gap:1px; padding-right:4px;">
              <div
                v-for="key in Object.keys(state.currentOpti)"
                :key="key"
                style="margin-bottom:8px;"
              >
                <label style="display:block; margin-bottom:4px; font-size:12px; color:#666;">
                  {{ key }}
                </label>
                <NInputNumber v-model:value="state.currentOpti[key as keyof typeof state.currentOpti]" :min="0" style="width: 100%;" />
              </div>
          </div>
            <div style="display:flex; flex-direction:column; gap:2px; margin-top:12px;">
              <NButton @click="evaluateCurrent" type="primary" style="width:100%;">Рассчитать</NButton>
            </div>    
        </NCard>

         <!-- Результаты оптимизации -->
        <NCard title="Результаты оптимизации" size="small" style="width:250px; height:380px;">
          <div style="height:200px; overflow-y:auto; display:flex; flex-direction:column; gap:1px; padding-right:4px;">
              <div
                v-for="key in Object.keys(state.optimization)"
                :key="key"
                style="margin-bottom:8px;"
              >
                <label style="display:block; margin-bottom:4px; font-size:12px; color:#666;">
                  {{ key }}
                </label>
                <NInputNumber v-model:value="state.optimization[key as keyof typeof state.optimization]" :min="0" style="width: 100%;" />
              </div>
          </div>
            <div style="display:flex; flex-direction:column; gap:2px; margin-top:12px;">
              <NButton @click="optimizeRSI" type="primary" style="width:100%;">Рассчитать</NButton>
            </div>    
        </NCard>

      </div>
    </NCard>

    
    <div 
      id="chart" 
      ref="chartEl" 
      style="
        width: 1104px; 
        max-width: 100%;
        min-height: 400px; 
        background: #f5f5f5;
        border-radius: 8px;
      "
    ></div>
  </div>
</template>

<style scoped>
</style>