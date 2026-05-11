// web/src/components/ChildSwitcher.tsx
import { useState, useEffect, useCallback } from 'react'
import { listChildren, createChild, deleteChild } from '../api/children'
import type { Child } from '../api/children'
import dayjs from 'dayjs'

interface Props {
  selectedId: string | null
  onSelect: (child: Child) => void
}

function ageLabel(birthDate: string): string {
  const months = Math.floor(dayjs().diff(dayjs(birthDate), 'day') / 30.4375)
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const rem = months % 12
  return rem > 0 ? `${years}岁${rem}个月` : `${years}岁`
}

export default function ChildSwitcher({ selectedId, onSelect }: Props) {
  const [children, setChildren] = useState<Child[]>([])
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [gender, setGender] = useState('male')
  const [birthDate, setBirthDate] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(async () => {
    try {
      const cs = await listChildren()
      setChildren(cs)
      if (cs.length > 0 && !selectedId) {
        onSelect(cs[0])
      }
    } catch { /* ignore */ }
  }, [selectedId, onSelect])

  useEffect(() => { load() }, [load])

  const handleCreate = async () => {
    if (!name || !birthDate) return
    setSubmitting(true)
    try {
      const child = await createChild({ name, gender, birth_date: birthDate })
      setChildren(prev => [...prev, child])
      onSelect(child)
      setShowForm(false)
      setName('')
      setBirthDate('')
    } catch { /* error */ }
    finally { setSubmitting(false) }
  }

  const handleDelete = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm('确认删除？将同时删除所有测量记录。')) return
    try {
      await deleteChild(id)
      const newChildren = children.filter(c => c.id !== id)
      setChildren(newChildren)
      if (selectedId === id && newChildren.length > 0) {
        onSelect(newChildren[0])
      } else if (newChildren.length === 0) {
        onSelect(null as any)
      }
    } catch { /* error */ }
  }

  return (
    <div>
      {/* Pills row */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        {children.map(child => (
          <div
            key={child.id}
            onClick={() => onSelect(child)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              padding: '6px 14px 6px 6px',
              borderRadius: 24,
              background: child.id === selectedId ? '#f0fdf4' : '#fff',
              border: child.id === selectedId ? '1.5px solid #4ade80' : '1.5px solid #e2e8f0',
              cursor: 'pointer',
              transition: 'all .15s',
              boxShadow: child.id === selectedId ? '0 0 0 3px rgba(34,197,94,.1)' : '0 1px 2px rgba(0,0,0,.04)',
            }}
          >
            <div style={{
              width: 32, height: 32, borderRadius: '50%',
              background: child.gender === 'male' ? '#dcfce7' : '#fff1f2',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 16, flexShrink: 0,
            }}>
              {child.gender === 'male' ? '👦' : '👧'}
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.3 }}>
              <div style={{ fontSize: 13, fontWeight: 600, color: child.id === selectedId ? '#15803d' : '#475569' }}>
                {child.name}
              </div>
              <div style={{ fontSize: 11, color: child.id === selectedId ? 'rgba(22,163,74,.75)' : '#94a3b8' }}>
                {ageLabel(child.birth_date)} · {child.gender === 'male' ? '男孩' : '女孩'} · {child.birth_date.split('T')[0]}
              </div>
            </div>
          </div>
        ))}

        {/* Add button / inline form */}
        {showForm ? (
          <div style={{
            padding: '6px 12px', borderRadius: 12, background: '#fff',
            border: '1.5px solid #4ade80', boxShadow: '0 0 0 3px rgba(34,197,94,.1)',
            display: 'flex', alignItems: 'center', gap: 6,
          }}>
            <input
              placeholder="姓名"
              value={name}
              onChange={e => setName(e.target.value)}
              style={{
                border: '1px solid #e2e8f0', borderRadius: 6, padding: '4px 8px',
                fontSize: 13, width: 90, outline: 'none', fontFamily: 'inherit',
              }}
              autoFocus
            />
            <select
              value={gender}
              onChange={e => setGender(e.target.value)}
              style={{
                border: '1px solid #e2e8f0', borderRadius: 6, padding: '4px 6px',
                fontSize: 13, outline: 'none', fontFamily: 'inherit', background: '#fff',
              }}
            >
              <option value="male">👦 男</option>
              <option value="female">👧 女</option>
            </select>
            <input
              type="date"
              value={birthDate}
              onChange={e => setBirthDate(e.target.value)}
              style={{
                border: '1px solid #e2e8f0', borderRadius: 6, padding: '4px 6px',
                fontSize: 13, outline: 'none', fontFamily: 'inherit',
              }}
            />
            <button
              onClick={handleCreate}
              disabled={submitting || !name || !birthDate}
              style={{
                background: '#22c55e', color: '#fff', border: 'none', borderRadius: 6,
                padding: '4px 14px', fontSize: 13, fontWeight: 600, cursor: submitting ? 'wait' : 'pointer',
                fontFamily: 'inherit', whiteSpace: 'nowrap',
              }}
            >
              {submitting ? '保存中...' : '保存'}
            </button>
            <button
              onClick={() => setShowForm(false)}
              style={{
                background: 'transparent', color: '#94a3b8', border: 'none', borderRadius: 6,
                padding: '4px 8px', fontSize: 13, cursor: 'pointer', fontFamily: 'inherit', whiteSpace: 'nowrap',
              }}
            >
              取消
            </button>
          </div>
        ) : (
          <div
            onClick={() => setShowForm(true)}
            style={{
              display: 'flex', alignItems: 'center', gap: 4,
              padding: '6px 12px', borderRadius: 24,
              border: '1.5px dashed #cbd5e1', background: 'transparent',
              cursor: 'pointer',
            }}
          >
            <span style={{ fontSize: 18, lineHeight: 1, color: '#94a3b8' }}>+</span>
          </div>
        )}
      </div>

      {/* Management row */}
      {children.length > 1 && (
        <div style={{ marginTop: 8, display: 'flex', gap: 6, alignItems: 'center' }}>
          <span style={{ fontSize: 11, color: '#94a3b8' }}>管理：</span>
          {children.map(child => (
            <div key={child.id} style={{
              display: 'flex', alignItems: 'center', gap: 4,
              padding: '2px 8px 2px 4px', borderRadius: 6,
              background: '#fff', border: '1px solid #f1f5f9',
            }}>
              <span style={{ fontSize: 12, color: '#64748b', fontWeight: 500 }}>{child.name}</span>
              <button
                onClick={(e) => handleDelete(child.id, e)}
                style={{
                  background: 'transparent', border: 'none', color: '#fb7185',
                  fontSize: 11, cursor: 'pointer', padding: '0 2px', lineHeight: 1,
                }}
              >
                删除
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
