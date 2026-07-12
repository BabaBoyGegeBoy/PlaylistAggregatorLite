<!-- 文件导入编辑：导入本工具导出的 txt/json，形成可勾选/编辑/增删/排序的列表，再导出 -->
<template>
  <div class="editor-tab">

    <!-- 文件选择 -->
    <div class="center-col">
      <div class="content-col">
        <label class="file-label btn button-center">
          {{ labels.chooseFile }}
          <input type="file" accept=".txt,.json" class="file-input" @change="onFile">
        </label>
        <p class="editor-hint">{{ labels.hint }}</p>
      </div>
    </div>

    <!-- 工具栏（有内容时显示） -->
    <div class="center-col" v-if="items.length">
      <div class="content-col editor-toolbar">
        <button class="btn" @click="toggleAll(true)">{{ labels.selectAll }}</button>
        <button class="btn" @click="toggleAll(false)">{{ labels.deselectAll }}</button>
        <button class="btn btn-primary" @click="addRow">{{ labels.addRow }}</button>
        <span class="editor-count">{{ labels.checked }}: {{ checkedCount }} / {{ items.length }}</span>
      </div>
    </div>

    <!-- 编辑表格 -->
    <div class="center-col" v-if="items.length">
      <div class="content-col">
        <table class="editor-table">
          <thead>
            <tr>
              <th class="col-check"></th>
              <th>{{ labels.colName }}</th>
              <th>{{ labels.colArtist }}</th>
              <th class="col-ops">{{ labels.colOps }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(it, idx) in items" :key="it.id">
              <td class="col-check">
                <input type="checkbox" v-model="it.checked">
              </td>
              <td>
                <input class="text-input editor-cell" v-model="it.name" :placeholder="labels.namePh">
              </td>
              <td>
                <input class="text-input editor-cell" v-model="it.artist" :placeholder="labels.artistPh">
              </td>
              <td class="col-ops">
                <button class="btn btn-mini" :disabled="idx === 0" @click="moveRow(idx, -1)" :title="labels.moveUp">↑</button>
                <button class="btn btn-mini" :disabled="idx === items.length - 1" @click="moveRow(idx, 1)" :title="labels.moveDown">↓</button>
                <button class="btn btn-mini btn-danger" @click="removeRow(idx)" :title="labels.remove">{{ labels.remove }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 空状态 -->
    <div class="center-col" v-else>
      <div class="content-col editor-empty">{{ labels.empty }}</div>
    </div>

    <!-- 导出 -->
    <div class="center-col" v-if="items.length">
      <div class="content-col editor-export">
        <button class="btn button-center" @click="exportTxt">{{ labels.exportTxt }}</button>
        <button class="btn btn-primary button-center" @click="exportJson">{{ labels.exportJson }}</button>
      </div>
    </div>

  </div>
</template>

<script setup>
import {reactive, computed} from 'vue';
import {sendErrorMessage, sendSuccessMessage} from "@/utils/tip";

// 列表项：id 用于 key；checked 是否勾选；name/artist 可编辑
const items = reactive([]);
let uid = 0;

const labels = {
  chooseFile: '选择 txt / json 文件',
  hint: '支持本工具导出的 txt（每行「歌名 - 歌手」）或 json（结构化 SongItem 数组 name + artists）。解析后可勾选、修改、增删、排序，再导出。',
  selectAll: '全选',
  deselectAll: '全不选',
  addRow: '新增一行',
  checked: '已选',
  colName: '歌名',
  colArtist: '歌手',
  colOps: '操作',
  namePh: '歌名',
  artistPh: '歌手（多歌手用、分隔）',
  empty: '请先选择文件，或点击「新增一行」手动添加。',
  exportTxt: '导出 .txt',
  exportJson: '导出 .json',
  moveUp: '上移',
  moveDown: '下移',
  remove: '删除',
};

const checkedCount = computed(() => items.filter(i => i.checked).length);

// 按最后一个 " - " 拆分歌名与歌手
function splitLine(line) {
  const idx = line.lastIndexOf(' - ');
  if (idx === -1) return {name: line.trim(), artist: ''};
  return {name: line.slice(0, idx).trim(), artist: line.slice(idx + 3).trim()};
}

function parseTxt(text) {
  return text.split(/\r?\n/)
    .map(l => l.trim())
    .filter(Boolean)
    .map(splitLine);
}

function parseJson(text) {
  const data = JSON.parse(text);
  let arr = [];
  if (Array.isArray(data)) arr = data;
  else if (data && Array.isArray(data.songs_detail)) arr = data.songs_detail;
  else if (data && Array.isArray(data.songs)) arr = data.songs;
  else throw new Error('无法识别的 JSON 结构（需为 SongItem 数组，或含 songs_detail/songs 字段的对象）');
  return arr.map(it => {
    if (typeof it === 'string') return splitLine(it);
    const name = it && it.name ? String(it.name) : '';
    let artist = '';
    if (it) {
      if (Array.isArray(it.artists)) artist = it.artists.join('、');
      else if (Array.isArray(it.artist)) artist = it.artist.join('、');
      else if (typeof it.artist === 'string') artist = it.artist;
    }
    return {name: name.trim(), artist: String(artist).trim()};
  });
}

function onFile(e) {
  const file = e.target.files && e.target.files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    try {
      const ext = (file.name.split('.').pop() || '').toLowerCase();
      const list = ext === 'json' ? parseJson(String(reader.result)) : parseTxt(String(reader.result));
      if (!list.length) {
        sendErrorMessage('文件为空或没有可解析的内容');
        return;
      }
      items.splice(0, items.length,
        ...list.map(x => ({id: ++uid, checked: true, name: x.name || '', artist: x.artist || ''})));
      sendSuccessMessage(`已导入 ${list.length} 首`);
    } catch (err) {
      sendErrorMessage('解析失败：' + (err.message || err));
    }
  };
  reader.onerror = () => sendErrorMessage('文件读取失败');
  reader.readAsText(file, 'utf-8');
  e.target.value = ''; // 允许重复选择同一文件
}

function addRow() {
  items.push({id: ++uid, checked: true, name: '', artist: ''});
}

function removeRow(idx) {
  items.splice(idx, 1);
}

function moveRow(idx, dir) {
  const j = idx + dir;
  if (j < 0 || j >= items.length) return;
  const tmp = items[idx];
  items[idx] = items[j];
  items[j] = tmp;
}

function toggleAll(val) {
  items.forEach(i => i.checked = val);
}

function splitArtist(s) {
  return String(s).split(/[、,/]/).map(x => x.trim()).filter(Boolean);
}

function download(filename, content, mime) {
  const blob = new Blob([content], {type: mime});
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

function exportTxt() {
  if (!items.length) {
    sendErrorMessage('列表为空');
    return;
  }
  const sel = items.filter(i => i.checked);
  if (!sel.length) {
    sendErrorMessage('请至少勾选一首');
    return;
  }
  const lines = sel.map(i => i.artist ? `${i.name} - ${i.artist}` : i.name);
  download('playlist.txt', lines.join('\n'), 'text/plain;charset=utf-8');
  sendSuccessMessage('已开始下载');
}

function exportJson() {
  if (!items.length) {
    sendErrorMessage('列表为空');
    return;
  }
  const sel = items.filter(i => i.checked);
  if (!sel.length) {
    sendErrorMessage('请至少勾选一首');
    return;
  }
  const arr = sel.map(i => ({name: i.name, artists: splitArtist(i.artist)}));
  download('playlist.json', JSON.stringify(arr, null, 2), 'application/json;charset=utf-8');
  sendSuccessMessage('已开始下载');
}
</script>

<style>
.editor-tab {
  margin-top: 4px;
}

.file-label {
  position: relative;
  overflow: hidden;
  cursor: pointer;
}

.file-input {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

.editor-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 6px;
  line-height: 1.6;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin: 4px 0;
}

.editor-count {
  font-size: 13px;
  color: #606266;
  margin-left: auto;
}

.editor-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  color: #303133;
}

.editor-table th,
.editor-table td {
  border: 1px solid #ebeef5;
  padding: 6px 8px;
  text-align: left;
  vertical-align: middle;
}

.editor-table th {
  background: #f5f7fa;
  color: #909399;
  font-weight: 600;
}

.editor-table tbody tr:nth-child(even) {
  background: #fafafa;
}

.editor-table .col-check {
  width: 42px;
  text-align: center;
}

.editor-table .col-ops {
  width: 132px;
  white-space: nowrap;
  text-align: center;
}

.editor-cell {
  margin: 0;
}

.btn-mini {
  padding: 4px 8px;
  min-width: 0;
  margin: 0 2px;
  font-size: 12px;
}

.editor-empty {
  color: #909399;
  padding: 20px;
  text-align: center;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
}

.editor-export {
  display: flex;
  gap: 10px;
}
</style>
