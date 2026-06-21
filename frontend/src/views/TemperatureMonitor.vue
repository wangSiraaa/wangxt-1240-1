<template>
  <div class="page">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span style="font-weight: 600; margin-right: 16px">粮温总览</span>
            <el-select v-model="selectedGranary" placeholder="选择仓房" clearable style="width: 200px" @change="reloadCharts">
              <el-option v-for="g in granaries" :key="g.id" :label="`${g.code} - ${g.name}`" :value="g.id" />
            </el-select>
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              style="width: 380px"
              value-format="YYYY-MM-DDTHH:mm:ss"
              @change="reloadCharts"
            />
          </div>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :span="6">
          <div class="gauge-block">
            <div class="gauge-title">最高温度</div>
            <v-chart :option="gaugeMaxOption" :autoresize="true" style="height: 180px" />
          </div>
        </el-col>
        <el-col :span="6">
          <div class="gauge-block">
            <div class="gauge-title">平均温度</div>
            <v-chart :option="gaugeAvgOption" :autoresize="true" style="height: 180px" />
          </div>
        </el-col>
        <el-col :span="6">
          <div class="gauge-block">
            <div class="gauge-title">最低温度</div>
            <v-chart :option="gaugeMinOption" :autoresize="true" style="height: 180px" />
          </div>
        </el-col>
        <el-col :span="6">
          <div class="gauge-block">
            <div class="gauge-title">平均湿度</div>
            <v-chart :option="gaugeHumidityOption" :autoresize="true" style="height: 180px" />
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="14">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>温度趋势曲线（所有传感器）</span>
              <el-tag size="small">实时</el-tag>
            </div>
          </template>
          <v-chart :option="trendOption" :autoresize="true" style="height: 380px" />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>仓房温度对比</span>
            </div>
          </template>
          <v-chart :option="compareOption" :autoresize="true" style="height: 380px" />
        </el-card>
      </el-col>
    </el-row>

    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>各仓房温度分布热力图</span>
            </div>
          </template>
          <v-chart :option="granaryHeatmapOption" :autoresize="true" style="height: 360px" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import type { Granary, SensorReading } from '@/types'
import { granaryApi } from '@/api'
import dayjs from 'dayjs'

const granaries = ref<Granary[]>([])
const selectedGranary = ref('')
const dateRange = ref<[string, string]>([
  dayjs().subtract(1, 'day').format('YYYY-MM-DDTHH:mm:ss'),
  dayjs().format('YYYY-MM-DDTHH:mm:ss')
])
const readings = ref<SensorReading[]>([])

const stats = reactive({
  maxTemp: 25.6,
  avgTemp: 18.4,
  minTemp: 12.8,
  avgHumidity: 62.5
})

const GranaryStatusLabels = {} as any

const loadGranaries = async () => {
  try {
    granaries.value = await granaryApi.list()
  } catch {
    granaries.value = []
  }
}

const reloadCharts = async () => {
  try {
    const gid = selectedGranary.value || granaries.value[0]?.id
    if (gid) {
      readings.value = await granaryApi.getReadings(gid, {
        start_time: dateRange.value?.[0],
        end_time: dateRange.value?.[1],
        limit: 2000
      })
    }
  } catch {
    readings.value = []
  }
  updateStats()
}

const updateStats = () => {
  const tempReadings = readings.value.filter(r => r.sensor_type === 'temperature')
  const humReadings = readings.value.filter(r => r.sensor_type === 'humidity')
  if (tempReadings.length) {
    const values = tempReadings.map(r => r.value)
    stats.maxTemp = Math.max(...values)
    stats.minTemp = Math.min(...values)
    stats.avgTemp = values.reduce((a, b) => a + b, 0) / values.length
  }
  if (humReadings.length) {
    stats.avgHumidity = humReadings.reduce((a, b) => a + b.value, 0) / humReadings.length
  }
}

const buildGaugeOption = (value: number, max: number, unit: string, label: string) => ({
  series: [{
    type: 'gauge',
    startAngle: 210, endAngle: -30,
    min: 0, max,
    progress: { show: true, width: 14 },
    axisLine: { lineStyle: { width: 14, color: [[0.5, '#67c23a'], [0.75, '#e6a23c'], [1, '#f56c6c']] } },
    pointer: { width: 5, length: '60%' },
    axisTick: { show: false },
    splitLine: { length: 8, lineStyle: { width: 2 } },
    axisLabel: { distance: 18, fontSize: 10 },
    anchor: { show: true, showAbove: true, size: 12, itemStyle: { borderWidth: 8 } },
    title: { show: false },
    detail: {
      valueAnimation: true,
      fontSize: 24,
      fontWeight: 700,
      offsetCenter: [0, '30%'],
      formatter: `{value} ${unit}`
    },
    data: [{ value: Number(value.toFixed(1)), name: label }]
  }]
})

const gaugeMaxOption = computed(() => buildGaugeOption(stats.maxTemp, 40, '°C', '最高温度'))
const gaugeAvgOption = computed(() => buildGaugeOption(stats.avgTemp, 40, '°C', '平均温度'))
const gaugeMinOption = computed(() => buildGaugeOption(stats.minTemp, 40, '°C', '最低温度'))
const gaugeHumidityOption = computed(() => buildGaugeOption(stats.avgHumidity, 100, '%', '平均湿度'))

const trendOption = computed(() => {
  const bySensor: Record<string, any[]> = {}
  const sorted = [...readings.value]
    .filter(r => r.sensor_type === 'temperature')
    .sort((a, b) => new Date(a.reading_time).getTime() - new Date(b.reading_time).getTime())
  for (const r of sorted) {
    const key = `${r.sensor_code} (${r.location_desc})`
    if (!bySensor[key]) bySensor[key] = []
    bySensor[key].push([dayjs(r.reading_time).format('MM-DD HH:mm'), r.value])
  }
  const colors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#8e44ad', '#16a085']
  let idx = 0
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: Object.keys(bySensor), top: 0 },
    grid: { left: '3%', right: '3%', bottom: '12%', top: '14%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false },
    yAxis: { type: 'value', name: '温度(°C)', scale: true },
    dataZoom: [{ type: 'inside' }, { type: 'slider', height: 20, bottom: 10 }],
    series: Object.entries(bySensor).map(([name, data]) => ({
      name, type: 'line', smooth: true, showSymbol: false, data,
      itemStyle: { color: colors[idx++ % colors.length] },
      lineStyle: { width: 2 }
    }))
  }
})

const compareOption = computed(() => {
  const names = granaries.value.map(g => `${g.code}`)
  const tempReadings = readings.value.filter(r => r.sensor_type === 'temperature')
  const statsByGranary = granaries.value.map(g => {
    const gReadings = tempReadings.filter(r => r.granary_id === g.id || g.id === (selectedGranary.value || granaries.value[0]?.id))
    if (gReadings.length === 0) return { avg: 0, max: 0, min: 0 }
    const vals = gReadings.map(r => r.value)
    return {
      avg: Number((vals.reduce((a, b) => a + b, 0) / vals.length).toFixed(1)),
      max: Number(Math.max(...vals).toFixed(1)),
      min: Number(Math.min(...vals).toFixed(1))
    }
  })
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['最高温', '平均温', '最低温'], top: 0 },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '14%', containLabel: true },
    xAxis: { type: 'category', data: names },
    yAxis: { type: 'value', name: '温度(°C)' },
    series: [
      { name: '最高温', type: 'bar', data: statsByGranary.map(s => s.max), itemStyle: { color: '#f56c6c' } },
      { name: '平均温', type: 'bar', data: statsByGranary.map(s => s.avg), itemStyle: { color: '#409eff' } },
      { name: '最低温', type: 'bar', data: statsByGranary.map(s => s.min), itemStyle: { color: '#67c23a' } }
    ]
  }
})

const granaryHeatmapOption = computed(() => {
  const hours = Array.from({ length: 24 }, (_, i) => `${i}:00`)
  const granaryNames = granaries.value.map(g => g.name)
  const data: any[] = []
  const tempReadings = readings.value.filter(r => r.sensor_type === 'temperature')
  for (let y = 0; y < granaryNames.length; y++) {
    for (let x = 0; x < hours.length; x++) {
      const gId = granaries.value[y]?.id
      const hourReadings = tempReadings.filter(r => {
        if (r.granary_id !== gId && gId !== (selectedGranary.value || granaries.value[0]?.id)) return false
        return new Date(r.reading_time).getHours() === x
      })
      const val = hourReadings.length > 0
        ? Number((hourReadings.reduce((sum, r) => sum + r.value, 0) / hourReadings.length).toFixed(1))
        : 0
      data.push([x, y, val])
    }
  }
  return {
    tooltip: { formatter: (p: any) => `${granaryNames[p.data[1]]} ${hours[p.data[0]]}<br/>温度: <b>${p.data[2]} °C</b>` },
    grid: { height: '60%', top: '8%' },
    xAxis: { type: 'category', data: hours, splitArea: { show: true }, axisLabel: { interval: 2 } },
    yAxis: { type: 'category', data: granaryNames, splitArea: { show: true } },
    visualMap: {
      min: 10, max: 35, calculable: true, orient: 'horizontal', left: 'center', bottom: '2%',
      inRange: { color: ['#2ecc71', '#3498db', '#f1c40f', '#e67e22', '#e74c3c'] }
    },
    series: [{
      name: '温度分布',
      type: 'heatmap',
      data,
      label: { show: true, fontSize: 9, color: '#fff' },
      emphasis: { itemStyle: { shadowBlur: 10 } }
    }]
  }
})

onMounted(() => {
  loadGranaries().then(() => reloadCharts())
})
</script>

<style scoped>
.page { width: 100%; }
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.gauge-block {
  background: linear-gradient(135deg, #f8f9ff, #f0f4ff);
  border-radius: 8px;
  padding: 12px;
}
.gauge-title {
  text-align: center;
  font-size: 13px;
  color: #606266;
  margin-bottom: -8px;
}
</style>
