<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea, #764ba2)">
              <el-icon :size="28"><OfficeBuilding /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.total_granaries }}</div>
              <div class="stat-label">仓房总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #11998e, #38ef7d)">
              <el-icon :size="28"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.normal_granaries }}</div>
              <div class="stat-label">正常仓房</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb, #f5576c)">
              <el-icon :size="28"><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.fumigating_granaries }}</div>
              <div class="stat-label">熏蒸中</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe, #00f2fe)">
              <el-icon :size="28"><RefreshRight /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.ventilating_granaries }}</div>
              <div class="stat-label">通风中</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a, #fee140)">
              <el-icon :size="28"><Bell /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value urgent">{{ stats.urgent_suggestions }}</div>
              <div class="stat-label">紧急建议</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #a18cd1, #fbc2eb)">
              <el-icon :size="28"><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.today_records }}</div>
              <div class="stat-label">今日记录</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span><el-icon><PieChart /></el-icon> 仓房状态分布</span>
            </div>
          </template>
          <v-chart :option="statusChartOption" :autoresize="true" style="height: 320px" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span><el-icon><TrendCharts /></el-icon> 近7日平均粮温趋势</span>
            </div>
          </template>
          <v-chart :option="tempTrendOption" :autoresize="true" style="height: 320px" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span><el-icon><Document /></el-icon> 翻仓建议统计</span>
              <el-button type="primary" size="small" link @click="$router.push('/turnover-suggestions')">查看全部</el-button>
            </div>
          </template>
          <v-chart :option="suggestionChartOption" :autoresize="true" style="height: 300px" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span><el-icon><WarningFilled /></el-icon> 熏蒸方案进度</span>
              <el-button type="primary" size="small" link @click="$router.push('/fumigation')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="fumigationList.slice(0, 5)" size="small" style="width: 100%">
            <el-table-column prop="plan_no" label="方案编号" width="150" />
            <el-table-column prop="plan_title" label="方案名称" show-overflow-tooltip />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="FumigationStatusColors[row.status] as any" size="small">
                  {{ FumigationStatusLabels[row.status] }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span><el-icon><List /></el-icon> 今日预警</span>
            </div>
          </template>
          <el-empty v-if="urgentSuggestions.length === 0" description="暂无预警信息" :image-size="80" />
          <el-table v-else :data="urgentSuggestions" size="small">
            <el-table-column prop="suggestion_no" label="建议编号" width="160" />
            <el-table-column label="仓房" width="120">
              <template #default="{ row }">{{ row.granary?.name || row.granary?.code || '-' }}</template>
            </el-table-column>
            <el-table-column prop="abnormal_area_desc" label="异常描述" show-overflow-tooltip />
            <el-table-column label="优先级" width="90">
              <template #default="{ row }">
                <el-tag :type="PriorityColors[row.priority] as any" size="small">
                  {{ PriorityLabels[row.priority] }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="生成时间" width="180">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="$router.push(`/turnover-suggestions`)">处理</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  OfficeBuilding,
  CircleCheck,
  Warning,
  RefreshRight,
  Bell,
  Clock,
  PieChart,
  TrendCharts,
  Document,
  WarningFilled,
  List
} from '@element-plus/icons-vue'
import type { DashboardStats, FumigationPlan, GrainTurnoverSuggestion, GrainConditionRecord } from '@/types'
import { FumigationStatusLabels, FumigationStatusColors, PriorityLabels, PriorityColors } from '@/types'
import { dashboardApi, fumigationApi, suggestionApi, grainConditionApi } from '@/api'
import dayjs from 'dayjs'

const stats = ref<DashboardStats>({
  total_granaries: 0,
  normal_granaries: 0,
  fumigating_granaries: 0,
  ventilating_granaries: 0,
  sealed_granaries: 0,
  abnormal_granaries: 0,
  pending_suggestions: 0,
  processing_suggestions: 0,
  urgent_suggestions: 0,
  high_suggestions: 0,
  pending_fumigation: 0,
  in_progress_fumigation: 0,
  today_records: 0,
  today_avg_temp: 0,
  today_max_temp: 0,
  abnormal_temp_count: 0
})

const tempRecords = ref<GrainConditionRecord[]>([])

const fumigationList = ref<FumigationPlan[]>([])
const urgentSuggestions = ref<GrainTurnoverSuggestion[]>([])

const loadData = async () => {
  try {
    stats.value = await dashboardApi.getStats()
  } catch {
  }
  try {
    const data = await fumigationApi.listPlans()
    fumigationList.value = data.slice(0, 5)
  } catch {
    fumigationList.value = []
  }
  try {
    const data = await suggestionApi.list({ status: 'pending' })
    urgentSuggestions.value = data.filter(s => s.priority === 'urgent' || s.priority === 'high')
  } catch {
    urgentSuggestions.value = []
  }
  try {
    const startDate = dayjs().subtract(6, 'day').format('YYYY-MM-DD')
    const endDate = dayjs().add(1, 'day').format('YYYY-MM-DD')
    tempRecords.value = await grainConditionApi.list({ start_date: startDate, end_date: endDate })
  } catch {
    tempRecords.value = []
  }
}

const formatTime = (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm')

const statusChartOption = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: '5%', left: 'center' },
  series: [{
    name: '仓房状态',
    type: 'pie',
    radius: ['40%', '70%'],
    avoidLabelOverlap: false,
    itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 2 },
    label: { show: false },
    emphasis: { label: { show: true, fontSize: 16, fontWeight: 'bold' } },
    data: [
      { value: stats.value.normal_granaries, name: '正常', itemStyle: { color: '#67c23a' } },
      { value: stats.value.fumigating_granaries, name: '熏蒸中', itemStyle: { color: '#f56c6c' } },
      { value: stats.value.ventilating_granaries, name: '通风中', itemStyle: { color: '#e6a23c' } },
      { value: stats.value.sealed_granaries, name: '已密封', itemStyle: { color: '#909399' } },
      { value: stats.value.abnormal_granaries, name: '异常', itemStyle: { color: '#f56c6c' } }
    ]
  }]
}))

const tempTrendOption = computed(() => {
  const days: string[] = []
  const avgTemps: (number | null)[] = []
  const maxTemps: (number | null)[] = []
  const minTemps: (number | null)[] = []

  for (let i = 6; i >= 0; i--) {
    const d = dayjs().subtract(i, 'day')
    const dayStr = d.format('YYYY-MM-DD')
    days.push(d.format('MM-DD'))

    const dayRecords = tempRecords.value.filter(r => dayjs(r.record_time).format('YYYY-MM-DD') === dayStr)
    if (dayRecords.length === 0) {
      avgTemps.push(null)
      maxTemps.push(null)
      minTemps.push(null)
    } else {
      const avg = dayRecords.reduce((s, r) => s + (r.avg_temperature || 0), 0) / dayRecords.length
      const max = Math.max(...dayRecords.map(r => r.max_temperature || 0))
      const min = Math.min(...dayRecords.map(r => r.min_temperature || 0))
      avgTemps.push(Number(avg.toFixed(1)))
      maxTemps.push(Number(max.toFixed(1)))
      minTemps.push(Number(min.toFixed(1)))
    }
  }

  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['平均温度', '最高温度', '最低温度'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: days },
    yAxis: { type: 'value', name: '温度(°C)' },
    series: [
      { name: '平均温度', type: 'line', smooth: true, data: avgTemps, areaStyle: { opacity: 0.1 }, itemStyle: { color: '#409eff' }, connectNulls: true },
      { name: '最高温度', type: 'line', smooth: true, data: maxTemps, itemStyle: { color: '#f56c6c' }, connectNulls: true },
      { name: '最低温度', type: 'line', smooth: true, data: minTemps, itemStyle: { color: '#67c23a' }, connectNulls: true }
    ]
  }
})

const suggestionChartOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['数量'] },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: { type: 'category', data: ['紧急建议', '高优先级', '待处理', '处理中'] },
  yAxis: { type: 'value', name: '数量' },
  series: [
    {
      name: '数量',
      type: 'bar',
      barWidth: '50%',
      data: [
        { value: stats.value.urgent_suggestions, itemStyle: { color: '#f56c6c' } },
        { value: stats.value.high_suggestions, itemStyle: { color: '#e6a23c' } },
        { value: stats.value.pending_suggestions, itemStyle: { color: '#409eff' } },
        { value: stats.value.processing_suggestions, itemStyle: { color: '#67c23a' } }
      ],
      label: { show: true, position: 'top' }
    }
  ]
}))

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  width: 100%;
}

.stat-card {
  margin-bottom: 0;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-value.urgent {
  color: #f56c6c;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
}

.card-header :deep(.el-icon) {
  margin-right: 6px;
  vertical-align: -2px;
}
</style>
