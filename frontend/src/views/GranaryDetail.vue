<template>
  <div class="page">
    <el-page-header @back="$router.back()" :content="detail?.name || '仓房详情'" class="page-header">
      <template #extra>
        <el-tag v-if="detail" :type="GranaryStatusColors[detail.status] as any" effect="dark" size="large">
          {{ GranaryStatusLabels[detail.status] }}
        </el-tag>
      </template>
    </el-page-header>

    <el-row :gutter="20" style="margin-top: 16px">
      <el-col :span="10">
        <el-card shadow="hover">
          <template #header><div class="card-title">基本信息</div></template>
          <el-descriptions v-if="detail" :column="2" border size="small">
            <el-descriptions-item label="仓房编号">{{ detail.code }}</el-descriptions-item>
            <el-descriptions-item label="仓房名称">{{ detail.name }}</el-descriptions-item>
            <el-descriptions-item label="位置" :span="2">{{ detail.location || '-' }}</el-descriptions-item>
            <el-descriptions-item label="粮食类型">{{ detail.grain_type || '-' }}</el-descriptions-item>
            <el-descriptions-item label="品种">{{ detail.grain_variety || '-' }}</el-descriptions-item>
            <el-descriptions-item label="容量(吨)">{{ detail.capacity || '-' }}</el-descriptions-item>
            <el-descriptions-item label="当前库存(吨)">{{ detail.grain_weight || 0 }}</el-descriptions-item>
            <el-descriptions-item label="保管员" :span="2">{{ detail.keeper?.full_name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="备注" :span="2">{{ detail.remark || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="hover" style="margin-top: 16px">
          <template #header>
            <div class="card-header">
              <span>传感器列表</span>
              <el-button v-if="userStore.hasRole('admin','keeper')" type="primary" size="small" :icon="Plus" @click="openSensorDialog">
                添加传感器
              </el-button>
            </div>
          </template>
          <el-table :data="sensors" size="small" empty-text="暂无传感器">
            <el-table-column prop="code" label="编号" width="110" />
            <el-table-column label="类型" width="110">
              <template #default="{ row }">{{ SensorTypeLabels[row.type] }}</template>
            </el-table-column>
            <el-table-column prop="location_desc" label="位置" show-overflow-tooltip />
            <el-table-column label="坐标" width="140">
              <template #default="{ row }">
                {{ row.position_x }},{{ row.position_y }},{{ row.position_z }}
              </template>
            </el-table-column>
            <el-table-column prop="unit" label="单位" width="70" />
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="14">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>粮温实时监测</span>
              <el-select v-model="sensorType" size="small" style="width: 120px" @change="reloadCharts">
                <el-option
                  v-for="(label, key) in SensorTypeLabels"
                  :key="key"
                  :label="label"
                  :value="key"
                />
              </el-select>
            </div>
          </template>
          <v-chart :option="tempChartOption" :autoresize="true" style="height: 340px" />
        </el-card>

        <el-card shadow="hover" style="margin-top: 16px">
          <template #header>
            <div class="card-title">粮温分布热力图（横切面）</div>
          </template>
          <v-chart :option="heatmapOption" :autoresize="true" style="height: 260px" />
        </el-card>

        <el-card shadow="hover" style="margin-top: 16px">
          <template #header>
            <div class="card-title">最近粮情记录</div>
          </template>
          <el-table :data="conditionRecords" size="small" empty-text="暂无记录">
            <el-table-column label="时间" width="170">
              <template #default="{ row }">{{ formatTime(row.record_time) }}</template>
            </el-table-column>
            <el-table-column label="均温(°C)" prop="avg_temperature" width="90" />
            <el-table-column label="最高(°C)" prop="max_temperature" width="90">
              <template #default="{ row }">
                <span :class="row.max_temperature > 25 ? 'temp-danger' : ''">{{ row.max_temperature }}</span>
              </template>
            </el-table-column>
            <el-table-column label="最低(°C)" prop="min_temperature" width="90" />
            <el-table-column label="湿度(%)" prop="avg_humidity" width="90" />
            <el-table-column label="异常">
              <template #default="{ row }">
                <el-tag v-if="row.pest_found" type="warning" size="small" effect="dark">虫害</el-tag>
                <el-tag v-if="row.mold_found" type="danger" size="small" effect="dark">霉变</el-tag>
                <span v-if="!row.pest_found && !row.mold_found">-</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="sensorDialogVisible" title="添加传感器" width="500px" destroy-on-close>
      <el-form :model="sensorForm" label-width="100px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="传感器编号"><el-input v-model="sensorForm.code" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型">
              <el-select v-model="sensorForm.type" style="width: 100%">
                <el-option
                  v-for="(label, key) in SensorTypeLabels"
                  :key="key"
                  :label="label"
                  :value="key"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="位置描述"><el-input v-model="sensorForm.location_desc" /></el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="X坐标"><el-input-number v-model="sensorForm.position_x" style="width: 100%" /></el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Y坐标"><el-input-number v-model="sensorForm.position_y" style="width: 100%" /></el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="Z坐标"><el-input-number v-model="sensorForm.position_z" style="width: 100%" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="单位"><el-input v-model="sensorForm.unit" /></el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="sensorDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddSensor">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import type { Granary, Sensor, SensorType, GrainConditionRecord, SensorReading } from '@/types'
import { GranaryStatusLabels, GranaryStatusColors, SensorTypeLabels } from '@/types'
import { granaryApi, grainConditionApi } from '@/api'
import dayjs from 'dayjs'

const route = useRoute()
const userStore = useUserStore()
const granaryId = computed(() => route.params.id as string)

const detail = ref<Granary>()
const sensors = ref<Sensor[]>([])
const conditionRecords = ref<GrainConditionRecord[]>([])
const readings = ref<SensorReading[]>([])
const sensorType = ref<SensorType>('temperature')

const sensorDialogVisible = ref(false)
const sensorForm = reactive({
  code: '', type: 'temperature' as SensorType, location_desc: '',
  position_x: 0, position_y: 0, position_z: 0, unit: '°C'
})

const GranaryStatusLabelsRef = GranaryStatusLabels
const GranaryStatusColorsRef = GranaryStatusColors
const SensorTypeLabelsRef = SensorTypeLabels

const loadDetail = async () => {
  try {
    detail.value = await granaryApi.get(granaryId.value)
  } catch {
  }
}

const loadSensors = async () => {
  try {
    sensors.value = await granaryApi.getSensors(granaryId.value)
  } catch {
    sensors.value = []
  }
}

const loadRecords = async () => {
  try {
    const data = await grainConditionApi.list({ granary_id: granaryId.value })
    conditionRecords.value = data.slice(0, 5)
  } catch {
    conditionRecords.value = []
  }
}

const loadReadings = async () => {
  try {
    readings.value = await granaryApi.getReadings(granaryId.value, { type: sensorType.value, limit: 500 })
  } catch {
    readings.value = []
  }
}

const reloadCharts = () => {
  loadReadings()
}

const formatTime = (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm')

const tempChartOption = computed(() => {
  const bySensor: Record<string, any[]> = {}
  const sorted = [...readings.value].sort((a, b) => new Date(a.reading_time).getTime() - new Date(b.reading_time).getTime())
  for (const r of sorted) {
    const key = r.sensor_code || r.sensor_id
    if (!bySensor[key]) bySensor[key] = []
    bySensor[key].push([dayjs(r.reading_time).format('HH:mm'), r.value])
  }
  const colors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#8e44ad']
  let idx = 0
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: Object.keys(bySensor) },
    grid: { left: '3%', right: '4%', bottom: '10%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false },
    yAxis: { type: 'value', name: SensorTypeLabels[sensorType.value] },
    dataZoom: [{ type: 'inside' }, { type: 'slider', height: 20, bottom: 5 }],
    series: Object.entries(bySensor).map(([name, data]) => ({
      name,
      type: 'line',
      smooth: true,
      showSymbol: false,
      data,
      itemStyle: { color: colors[idx++ % colors.length] },
      areaStyle: { opacity: 0.05 }
    }))
  }
})

const heatmapOption = computed(() => {
  const xSize = 10
  const ySize = 8

  const sensorMap: Record<string, Sensor> = {}
  for (const s of sensors.value) {
    sensorMap[s.id] = s
  }

  const latestBySensor: Record<string, SensorReading> = {}
  for (const r of readings.value) {
    const existing = latestBySensor[r.sensor_id]
    if (!existing || new Date(r.reading_time) > new Date(existing.reading_time)) {
      latestBySensor[r.sensor_id] = r
    }
  }

  const data: any[] = []
  let minVal = Infinity
  let maxVal = -Infinity
  for (const [sensorId, reading] of Object.entries(latestBySensor)) {
    const sensor = sensorMap[sensorId]
    if (!sensor) continue
    const px = sensor.position_x ?? 0
    const py = sensor.position_y ?? 0
    const x = Math.max(0, Math.min(xSize - 1, Math.round(px)))
    const y = Math.max(0, Math.min(ySize - 1, Math.round(py)))
    data.push([x, y, Number(reading.value.toFixed(1))])
    minVal = Math.min(minVal, reading.value)
    maxVal = Math.max(maxVal, reading.value)
  }

  if (data.length === 0) {
    minVal = 0
    maxVal = 1
  }

  return {
    tooltip: {
      position: 'top',
      formatter: (p: any) => `位置(${p.data[0]}, ${p.data[1]})<br/>${SensorTypeLabelsRef[sensorType.value]}: <b>${p.data[2]}</b>`
    },
    grid: { height: '60%', top: '10%' },
    xAxis: { type: 'category', data: Array.from({ length: xSize }, (_, i) => `X${i + 1}`), splitArea: { show: true } },
    yAxis: { type: 'category', data: Array.from({ length: ySize }, (_, i) => `Y${ySize - i}`), splitArea: { show: true } },
    visualMap: {
      min: Number.isFinite(minVal) ? Math.floor(minVal) : 0,
      max: Number.isFinite(maxVal) ? Math.ceil(maxVal) : 1,
      calculable: true, orient: 'horizontal', left: 'center', bottom: '5%',
      inRange: { color: ['#2ecc71', '#f1c40f', '#e67e22', '#e74c3c'] }
    },
    series: [{
      name: SensorTypeLabelsRef[sensorType.value],
      type: 'heatmap',
      data,
      label: { show: true, fontSize: 10 },
      emphasis: { itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0, 0, 0, 0.5)' } }
    }]
  }
})

const openSensorDialog = () => {
  Object.assign(sensorForm, { code: '', type: 'temperature', location_desc: '', position_x: 0, position_y: 0, position_z: 0, unit: '°C' })
  sensorDialogVisible.value = true
}

const handleAddSensor = async () => {
  try {
    await granaryApi.addSensor(granaryId.value, sensorForm)
    ElMessage.success('添加成功')
    sensorDialogVisible.value = false
    loadSensors()
  } catch {
  }
}

onMounted(() => {
  loadDetail()
  loadSensors()
  loadRecords()
  setTimeout(() => loadReadings(), 100)
})
</script>

<style scoped>
.page {
  width: 100%;
}

.page-header {
  background: #fff;
  padding: 16px 20px;
  border-radius: 4px;
  margin: -4px;
}

.card-title {
  font-weight: 600;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.temp-danger {
  color: #f56c6c;
  font-weight: 600;
}
</style>
