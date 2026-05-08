// web/src/components/DietForm.tsx
import { useEffect, useState } from 'react'
import {
  Drawer,
  Form,
  DatePicker,
  Input,
  Select,
  Radio,
  Button,
  message,
  Space,
  Card,
} from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import { createDiet } from '../api/diet'
import dayjs from 'dayjs'

const foodTypeOptions = [
  { value: 'staple', label: '主食' },
  { value: 'vegetable', label: '蔬菜' },
  { value: 'fruit', label: '水果' },
  { value: 'protein', label: '肉蛋' },
  { value: 'dairy', label: '奶' },
  { value: 'snack', label: '零食' },
]

interface Props {
  open: boolean
  childId: string
  onClose: () => void
  onSaved: () => void
}

interface DietItem {
  food_name: string
  food_type: string
  amount_level: number
  record_time: string
}

export default function DietForm({ open, childId, onClose, onSaved }: Props) {
  const [items, setItems] = useState<DietItem[]>([
    { food_name: '', food_type: 'staple', amount_level: 2, record_time: dayjs().format() },
  ])

  useEffect(() => {
    if (open && items.length === 0) {
      setItems([
        { food_name: '', food_type: 'staple', amount_level: 2, record_time: dayjs().format() },
      ])
    }
  }, [open])

  const updateItem = (index: number, field: string, value: any) => {
    setItems(prev => {
      const next = [...prev]
      next[index] = { ...next[index], [field]: value }
      return next
    })
  }

  const addItem = () => {
    setItems(prev => [
      ...prev,
      { food_name: '', food_type: 'staple', amount_level: 2, record_time: dayjs().format() },
    ])
  }

  const removeItem = (index: number) => {
    if (items.length <= 1) return
    setItems(prev => prev.filter((_, i) => i !== index))
  }

  const handleSave = async () => {
    const validItems = items.filter(item => item.food_name.trim())
    if (validItems.length === 0) {
      message.warning('请至少填写一种食物')
      return
    }
    try {
      await Promise.all(
        validItems.map(item =>
          createDiet(childId, {
            food_name: item.food_name.trim(),
            food_type: item.food_type,
            amount_level: item.amount_level,
            record_time: item.record_time,
          }),
        ),
      )
      onSaved()
      onClose()
      setItems([])
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '保存失败')
    }
  }

  const amountLabels = ['少吃了一点', '正常量', '吃了很多']

  return (
    <Drawer
      title="添加饮食记录"
      placement="right"
      open={open}
      onClose={onClose}
      width={420}
    >
      {items.map((item, index) => (
        <div key={index}>
          <Card
            size="small"
            title={`第 ${index + 1} 种`}
            extra={
              items.length > 1 && (
                <DeleteOutlined
                  style={{ color: '#ef4444', cursor: 'pointer' }}
                  onClick={() => removeItem(index)}
                />
              )
            }
            style={{ marginBottom: 12 }}
          >
            <Form layout="vertical" style={{ marginBottom: 12 }}>
              <Form.Item label="食物名称" required>
                <Input
                  value={item.food_name}
                  onChange={e => updateItem(index, 'food_name', e.target.value)}
                  placeholder="如：米粉、西兰花泥"
                />
              </Form.Item>
              <Form.Item label="食物类型">
                <Select
                  value={item.food_type}
                  onChange={v => updateItem(index, 'food_type', v)}
                  options={foodTypeOptions}
                />
              </Form.Item>
              <Form.Item label="食量">
                <Radio.Group
                  value={item.amount_level}
                  onChange={e => updateItem(index, 'amount_level', e.target.value)}
                >
                  <Space direction="vertical">
                    {[1, 2, 3].map(level => (
                      <Radio key={level} value={level}>
                        {amountLabels[level - 1]}
                      </Radio>
                    ))}
                  </Space>
                </Radio.Group>
              </Form.Item>
              <Form.Item label="时间">
                <DatePicker
                  showTime
                  value={dayjs(item.record_time)}
                  onChange={v =>
                    updateItem(index, 'record_time', v?.format() ?? dayjs().format())
                  }
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Form>
          </Card>
        </div>
      ))}
      <Button
        type="dashed"
        icon={<PlusOutlined />}
        block
        onClick={addItem}
        style={{ marginBottom: 16 }}
      >
        再加一条
      </Button>
      <Button
        type="primary"
        htmlType="button"
        block
        style={{ background: '#16a34a' }}
        onClick={handleSave}
      >
        保存 {items.filter(i => i.food_name.trim()).length} 条记录
      </Button>
    </Drawer>
  )
}
