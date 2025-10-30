import { useEffect, useMemo, useState } from 'react';
import useSWR from 'swr';
import { api } from '../lib/api';

interface DbTrader {
  trader_id: string;
  name: string;
  ai_model: 'qwen' | 'deepseek' | 'custom';
  exchange: 'binance' | 'hyperliquid' | 'aster';
  binance_api_key?: string;
  binance_secret_key?: string;
  binance_testnet?: boolean;
  qwen_key?: string;
  deepseek_key?: string;
  custom_api_url?: string;
  custom_api_key?: string;
  custom_model_name?: string;
  initial_balance: number;
  scan_interval_minutes: number;
  enabled: boolean;
  created_at?: string;
  updated_at?: string;
}

export default function AdminTraders() {
  const { data, mutate, isLoading } = useSWR<DbTrader[]>(`admin-traders`, api.adminListTraders, {
    refreshInterval: 10000,
  });

  const [editing, setEditing] = useState<DbTrader | null>(null);
  const [form, setForm] = useState<Partial<DbTrader>>({ enabled: true, ai_model: 'qwen', exchange: 'binance', scan_interval_minutes: 3, initial_balance: 5000 });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (editing) setForm(editing);
  }, [editing]);

  const onSave = async () => {
    if (!form.trader_id || !form.name || !form.ai_model || !form.exchange || !form.initial_balance || !form.scan_interval_minutes) {
      setError('请填写必填项');
      return;
    }
    setError(null);
    setSaving(true);
    try {
      await api.adminUpsertTrader(form);
      // 保存成功后重载交易者
      await api.adminReload();
      setEditing(null);
      setForm({ enabled: true, ai_model: 'qwen', exchange: 'binance', scan_interval_minutes: 3, initial_balance: 5000 });
      await mutate();
    } catch (e: any) {
      setError(e.message || '保存失败');
    } finally {
      setSaving(false);
    }
  };

  const onDelete = async (trader_id: string) => {
    if (!confirm(`确认删除 ${trader_id} ?`)) return;
    try {
      await api.adminDeleteTrader(trader_id);
      // 删除成功后重载交易者
      await api.adminReload();
      await mutate();
    } catch (e) {
      alert('删除失败');
    }
  };

  const onToggle = async (trader_id: string, enabled: boolean) => {
    try {
      await api.setTradingEnabled(trader_id, enabled);
      await mutate();
    } catch (e) {
      alert('切换失败');
    }
  };

  return (
    <div className="space-y-6">
      <div className="binance-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold" style={{ color: '#EAECEF' }}>Trader 管理</h2>
          <button
            className="px-3 py-2 rounded text-sm font-semibold"
            style={{ background: '#F0B90B', color: '#000' }}
            onClick={() => { setEditing(null); setForm({ enabled: true, ai_model: 'qwen', exchange: 'binance', scan_interval_minutes: 3, initial_balance: 5000 }); }}
          >新增 Trader</button>
        </div>

        {/* 列表 */}
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="text-left border-b border-gray-800">
              <tr>
                <th className="pb-3 text-gray-400">ID</th>
                <th className="pb-3 text-gray-400">名称</th>
                <th className="pb-3 text-gray-400">AI</th>
                <th className="pb-3 text-gray-400">交易所</th>
                <th className="pb-3 text-gray-400">初始资金</th>
                <th className="pb-3 text-gray-400">周期(分)</th>
                <th className="pb-3 text-gray-400">开关</th>
                <th className="pb-3 text-gray-400">操作</th>
              </tr>
            </thead>
            <tbody>
              {(data || []).map((row) => (
                <tr key={row.trader_id} className="border-b border-gray-800">
                  <td className="py-3 font-mono">{row.trader_id}</td>
                  <td className="py-3">{row.name}</td>
                  <td className="py-3">{row.ai_model?.toUpperCase()}</td>
                  <td className="py-3">{row.exchange}</td>
                  <td className="py-3">{row.initial_balance}</td>
                  <td className="py-3">{row.scan_interval_minutes}</td>
                  <td className="py-3">
                    <label className="inline-flex items-center gap-2 cursor-pointer">
                      <input type="checkbox" checked={!!row.enabled} onChange={(e) => onToggle(row.trader_id, e.target.checked)} />
                      <span style={{ color: row.enabled ? '#0ECB81' : '#F6465D' }}>{row.enabled ? '开启' : '关闭'}</span>
                    </label>
                  </td>
                  <td className="py-3 flex gap-2">
                    <button className="px-2 py-1 rounded text-xs" style={{ background: '#2B3139', color: '#EAECEF' }} onClick={() => { setEditing(row); setForm(row); }}>编辑</button>
                    <button className="px-2 py-1 rounded text-xs" style={{ background: 'rgba(246,70,93,.15)', color: '#F6465D' }} onClick={() => onDelete(row.trader_id)}>删除</button>
                  </td>
                </tr>
              ))}
              {(!data || data.length === 0) && (
                <tr><td className="py-6 text-center text-gray-500" colSpan={8}>暂无数据</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* 表单 */}
      <div className="binance-card p-6">
        <h3 className="text-lg font-bold mb-4" style={{ color: '#EAECEF' }}>{editing ? '编辑 Trader' : '新增 Trader'}</h3>
        {error && <div className="mb-3 text-sm px-3 py-2 rounded" style={{ background: 'rgba(246,70,93,.1)', color: '#F6465D' }}>{error}</div>}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input label="Trader ID" value={form.trader_id || ''} onChange={(v) => setForm({ ...form, trader_id: v })} disabled={!!editing} />
          <Input label="名称" value={form.name || ''} onChange={(v) => setForm({ ...form, name: v })} />
          <Select label="AI 模型" value={form.ai_model || 'qwen'} options={[['qwen','Qwen'],['deepseek','DeepSeek'],['custom','Custom']]} onChange={(v) => setForm({ ...form, ai_model: v as any })} />
          <Select label="交易所" value={form.exchange || 'binance'} options={[['binance','Binance'],['hyperliquid','Hyperliquid'],['aster','Aster']]} onChange={(v) => setForm({ ...form, exchange: v as any })} />
          <Input label="初始资金" type="number" value={String(form.initial_balance || '')} onChange={(v) => setForm({ ...form, initial_balance: Number(v) })} />
          <Input label="扫描周期(分钟)" type="number" value={String(form.scan_interval_minutes || '')} onChange={(v) => setForm({ ...form, scan_interval_minutes: Number(v) })} />
          <Checkbox label="币安测试网" checked={!!form.binance_testnet} onChange={(v) => setForm({ ...form, binance_testnet: v })} />
          <Checkbox label="启用交易" checked={!!form.enabled} onChange={(v) => setForm({ ...form, enabled: v })} />
          <Input label="Binance API Key" value={form.binance_api_key || ''} onChange={(v) => setForm({ ...form, binance_api_key: v })} />
          <Input label="Binance Secret Key" value={form.binance_secret_key || ''} onChange={(v) => setForm({ ...form, binance_secret_key: v })} />
          {form.ai_model === 'qwen' && (
            <Input label="Qwen Key" value={form.qwen_key || ''} onChange={(v) => setForm({ ...form, qwen_key: v })} />
          )}
          {form.ai_model === 'deepseek' && (
            <Input label="DeepSeek Key" value={form.deepseek_key || ''} onChange={(v) => setForm({ ...form, deepseek_key: v })} />
          )}
          {form.ai_model === 'custom' && (
            <>
              <Input label="Custom API URL" value={form.custom_api_url || ''} onChange={(v) => setForm({ ...form, custom_api_url: v })} />
              <Input label="Custom API Key" value={form.custom_api_key || ''} onChange={(v) => setForm({ ...form, custom_api_key: v })} />
              <Input label="Model Name" value={form.custom_model_name || ''} onChange={(v) => setForm({ ...form, custom_model_name: v })} />
            </>
          )}
        </div>
        <div className="mt-4 flex gap-3">
          <button className="px-4 py-2 rounded text-sm font-semibold" style={{ background: '#F0B90B', color: '#000' }} onClick={onSave} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
          <button className="px-4 py-2 rounded text-sm font-semibold" style={{ background: '#2B3139', color: '#EAECEF' }} onClick={() => { setEditing(null); setForm({ enabled: true, ai_model: 'qwen', exchange: 'binance', scan_interval_minutes: 3, initial_balance: 5000 }); }}>重置</button>
        </div>
      </div>
    </div>
  );
}

function Input({ label, value, onChange, type = 'text', disabled = false }: { label: string; value: string; onChange: (v: string) => void; type?: string; disabled?: boolean }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs" style={{ color: '#848E9C' }}>{label}</span>
      <input value={value} onChange={(e) => onChange(e.target.value)} type={type} disabled={disabled} className="px-3 py-2 rounded text-sm" style={{ background: '#1E2329', border: '1px solid #2B3139', color: '#EAECEF' }} />
    </label>
  );
}

function Select({ label, value, options, onChange }: { label: string; value: string; options: [string,string][]; onChange: (v: string) => void }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs" style={{ color: '#848E9C' }}>{label}</span>
      <select value={value} onChange={(e) => onChange(e.target.value)} className="px-3 py-2 rounded text-sm" style={{ background: '#1E2329', border: '1px solid #2B3139', color: '#EAECEF' }}>
        {options.map(([v, l]) => <option key={v} value={v}>{l}</option>)}
      </select>
    </label>
  );
}

function Checkbox({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center gap-2">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      <span style={{ color: '#EAECEF' }}>{label}</span>
    </label>
  );
}
