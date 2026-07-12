<!--  三段式：头部、主体、底部（轻量版：原生控件，无 element-plus） -->
<template>
  <div class="app-container">

    <header class="app-header">
      <p class="text-center title">{{ i18n.title }}</p>
      <p class="text-center subtitle">{{ i18n.subtitle }}</p>
    </header>

    <main class="app-main">

      <div class="mode-tabs">
        <!-- 页签切换（原生按钮） -->
        <div class="tab-bar">
          <button :class="['tab-btn', {active: state.activeTab === 'single'}]"
                  @click="state.activeTab = 'single'">{{ i18n.tabSingle }}</button>
          <button :class="['tab-btn', {active: state.activeTab === 'aggregate'}]"
                  @click="state.activeTab = 'aggregate'">{{ i18n.tabAggregate }}</button>
          <button :class="['tab-btn', {active: state.activeTab === 'editor'}]"
                  @click="state.activeTab = 'editor'">{{ i18n.tabEditor }}</button>
        </div>

        <!-- 单歌单 -->
        <div v-show="state.activeTab === 'single'">
          <div class="center-col">
            <div class="content-col">
              <input class="text-input" v-model="state.link"
                     :placeholder="i18n.inputPlaceholder"
                     @keyup.enter="fetchLinkDetails">
            </div>
          </div>

          <div class="compact-row center-col">
            <div class="content-col option-line">
              <label class="checkbox">
                <input type="checkbox" v-model="state.useDetailedSongName">
                <span>{{ i18n.detailedSongName }}</span>
              </label>
              <span class="info-icon" :title="i18n.detailedSongNameTip">ⓘ</span>
            </div>
          </div>

          <div class="center-col">
            <div class="content-col option-line">
              <span class="format-label">{{ i18n.songFormat }}:</span>
              <span class="format-radio-group">
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="song-singer"> {{ i18n.formatSongSinger }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="singer-song"> {{ i18n.formatSingerSong }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="song"> {{ i18n.formatSongOnly }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="custom"> {{ i18n.formatCustom }}</label>
              </span>
            </div>
          </div>

          <div class="center-col" v-if="state.songFormat === 'custom'">
            <div class="content-col custom-tmpl-block">
              <label class="tmpl-label">{{ i18n.customTmplTop }}</label>
              <textarea class="text-area" v-model="state.topTemplate" rows="3"></textarea>
              <label class="tmpl-label">{{ i18n.customTmplSong }}</label>
              <textarea class="text-area" v-model="state.songTemplate" rows="3"></textarea>
              <p class="custom-tmpl-tip">{{ i18n.customTmplTip }}</p>
            </div>
          </div>

          <div class="center-col">
            <div class="content-col option-line">
              <span class="format-label">{{ i18n.songOrder }}:</span>
              <span class="format-radio-group">
                <label class="radio-label"><input type="radio" v-model="state.songOrder" value="normal"> {{ i18n.orderNormal }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songOrder" value="reverse"> {{ i18n.orderReverse }}</label>
              </span>
            </div>
          </div>

          <button class="btn btn-danger button-center lang-song-list-btn" @click="throttledFetchLinkDetails">
            {{ i18n.fetchSongList }}
          </button>

          <div class="center-col">
            <div class="content-col">
              <textarea class="text-area" v-model="state.result" rows="15"
                        :placeholder="i18n.resultHint"></textarea>

              <div class="songs-count-display text-center" v-show="state.songsCount > 0">
                {{ i18n.songsCount }}: {{ state.songsCount }}
                <span class="platform-tag" v-if="state.platformName">{{ state.platformName }}</span>
              </div>
            </div>
          </div>

          <button class="btn button-center lang-copy-btn" @click="copyResult">
            {{ i18n.copy }}
          </button>
          <button class="btn button-center" v-if="state.songFormat === 'custom'" @click="downloadJson('playlist.json')">{{ i18n.downloadJson }}</button>
        </div>

        <!-- 多歌单聚合 -->
        <div v-show="state.activeTab === 'aggregate'">
          <div class="center-col">
            <div class="content-col">
              <textarea class="text-area" v-model="state.urls" rows="6"
                        :placeholder="i18n.aggInputPlaceholder"
                        @keyup.ctrl.enter="throttledAggregateLinkDetails"></textarea>
            </div>
          </div>

          <div class="compact-row center-col">
            <div class="content-col option-line">
              <label class="checkbox">
                <input type="checkbox" v-model="state.useDetailedSongName">
                <span>{{ i18n.detailedSongName }}</span>
              </label>
            </div>
          </div>

          <div class="center-col">
            <div class="content-col option-line">
              <span class="format-label">{{ i18n.songFormat }}:</span>
              <span class="format-radio-group">
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="song-singer"> {{ i18n.formatSongSinger }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="singer-song"> {{ i18n.formatSingerSong }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="song"> {{ i18n.formatSongOnly }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songFormat" value="custom"> {{ i18n.formatCustom }}</label>
              </span>
            </div>
          </div>

          <div class="center-col" v-if="state.songFormat === 'custom'">
            <div class="content-col custom-tmpl-block">
              <label class="tmpl-label">{{ i18n.customTmplTop }}</label>
              <textarea class="text-area" v-model="state.topTemplate" rows="3"></textarea>
              <label class="tmpl-label">{{ i18n.customTmplSong }}</label>
              <textarea class="text-area" v-model="state.songTemplate" rows="3"></textarea>
              <p class="custom-tmpl-tip">{{ i18n.customTmplTip }}</p>
            </div>
          </div>

          <div class="center-col">
            <div class="content-col option-line">
              <span class="format-label">{{ i18n.songOrder }}:</span>
              <span class="format-radio-group">
                <label class="radio-label"><input type="radio" v-model="state.songOrder" value="normal"> {{ i18n.orderNormal }}</label>
                <label class="radio-label"><input type="radio" v-model="state.songOrder" value="reverse"> {{ i18n.orderReverse }}</label>
              </span>
            </div>
          </div>

          <button class="btn btn-danger button-center" @click="throttledAggregateLinkDetails">
            {{ i18n.aggregateBtn }}
          </button>

          <div class="center-col">
            <div class="content-col">
              <textarea class="text-area" v-model="state.aggResult" rows="15"
                        :placeholder="i18n.aggResultHint"></textarea>

              <div class="songs-count-display text-center" v-show="state.aggCount > 0">
                {{ i18n.songsCount }}: {{ state.aggCount }}
                <span class="agg-dup-tag">{{ i18n.aggDuplicates }}: {{ state.aggDuplicates }}</span>
              </div>

              <div class="agg-source-wrap">
                <table class="agg-source-table" v-show="state.aggSources.length > 0">
                  <thead>
                    <tr>
                      <th>{{ i18n.aggSourcePlatform }}</th>
                      <th>{{ i18n.aggSourceCount }}</th>
                      <th>{{ i18n.aggSourceStatus }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in state.aggSources" :key="row.platform_name">
                      <td>{{ row.platform_name }}</td>
                      <td>{{ row.count }}</td>
                      <td>
                        <span v-if="row.ok" class="tag tag-ok">{{ i18n.aggSourceOK }}</span>
                        <span v-else class="tag tag-fail" :title="row.error">{{ i18n.aggSourceFail }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div class="center-col">
            <button class="btn button-center" @click="copyResult">{{ i18n.copy }}</button>
            <button class="btn btn-primary button-center" v-if="state.songFormat === 'custom'" @click="downloadJson('aggregate_playlist.json')">{{ i18n.downloadJson }}</button>
            <button class="btn btn-primary button-center" v-else @click="downloadAgg">{{ i18n.downloadTxt }}</button>
          </div>
        </div>

        <!-- 文件导入编辑 -->
        <EditorTab v-show="state.activeTab === 'editor'" />

      </div>

    </main>

    <!-- 右下角悬浮「?」帮助按钮 -->
    <button class="guide-fab" @click="guideDrawer = true" aria-label="使用指南">?</button>

    <!-- 使用指南抽屉 -->
    <div class="guide-drawer-mask" v-show="guideDrawer" @click.self="guideDrawer = false">
      <aside class="guide-drawer" :style="{ width: drawerSize }">
        <button class="guide-close" @click="guideDrawer = false" aria-label="关闭">×</button>
        <div class="guide-content">
          <h3>这是什么</h3>
          <p>歌单解析与聚合是一款<strong>本地运行</strong>的网页小工具：把各平台歌单链接，解析成统一的
            「歌名 - 歌手」文本，支持多歌单聚合去重，方便你迁移到其他音乐平台。</p>

          <h3>单歌单解析</h3>
          <ol>
            <li>切换到「单歌单解析」页签</li>
            <li>粘贴<strong>一个</strong>歌单链接（如 <code>http://163cn.tv/zoIxm3</code>）</li>
            <li>点击「获取歌单」，结果以「歌名 - 歌手」每行一首展示</li>
            <li>点击「复制结果」即可复制文本</li>
          </ol>

          <h3>多歌单聚合</h3>
          <ol>
            <li>切换到「多歌单聚合」页签</li>
            <li><strong>每行</strong>粘贴一个歌单链接（可填多个）</li>
            <li>点击「聚合歌单」，自动合并并去除重复歌曲，显示总数、去重数与各来源状态</li>
            <li>可「复制结果」或「下载 .txt」</li>
          </ol>

          <h3>文件导入编辑</h3>
          <ol>
            <li>切换到「文件导入编辑」页签</li>
            <li>点击「选择 txt / json 文件」，导入本工具导出的歌单：txt 每行「歌名 - 歌手」；json 为结构化 SongItem 数组（name + artists）</li>
            <li>在列表中可勾选、修改歌名/歌手、新增/删除行、用 ↑/↓ 调整顺序</li>
            <li>点击「导出 .txt」或「导出 .json」，把勾选的歌曲以浏览器下载方式重新导出（未勾选的不导出）</li>
          </ol>

          <h3>格式与顺序</h3>
          <ul>
            <li><strong>歌曲格式</strong>：歌名 - 歌手 / 歌手 - 歌名 / 仅歌名</li>
            <li><strong>歌曲顺序</strong>：正序 / 倒序</li>
            <li><strong>使用未经处理的原始歌曲名</strong>：默认不勾选。处理后的歌名在迁移到其他平台时匹配率更高；如需原样歌名可勾选。</li>
          </ul>

          <div v-pre>
          <h3>自定义 JSON 格式</h3>
          <p>除了上述文本格式，还可选择「<strong>自定义 JSON</strong>」输出任意结构的 JSON 文件，方便你直接对接程序或其它平台。</p>
          <p>采用<strong>双模板</strong>：先写「歌单模板」（顶层骨架），其中 <code>{{songs}}</code> 会被替换为所有单曲对象拼接而成；再写「单曲模板」（每首歌的对象）。可用占位符：</p>
          <ul>
            <li>歌单模板：<code>{{name}}</code>（歌单名）、<code>{{count}}</code>（歌曲数）、<code>{{songs}}</code>（单曲数组，需用 <code>[{{songs}}]</code> 包裹）</li>
            <li>单曲模板：<code>{{song.name}}</code>（歌名）、<code>{{song.artist}}</code>（歌手字符串，多歌手用 "、" 连接）、<code>{{song.artists}}</code>（歌手数组元素，可放进 <code>[...]</code>）、<code>{{song.album}}</code>（专辑）、<code>{{song.id}}</code>（歌曲 ID）</li>
          </ul>
          <p>默认模板即为下方示例，选「自定义 JSON」后可直接使用：</p>
          <pre class="guide-code">歌单模板：{"name":"{{name}}","tracks":[{{songs}}]}
单曲模板：{"name":"{{song.name}}","artist":[{{song.artists}}]}</pre>
          <p>渲染效果示例：</p>
          <pre class="guide-code">{
  "name": "歌单名称",
  "tracks": [
    { "name": "晴天", "artist": ["周杰伦"] },
    { "name": "浮夸", "artist": ["陈奕迅"] }
  ]
}</pre>
          </div>

          <h3>迁移到其他平台（第三方服务）</h3>
          <p>解析出的文本可直接用于迁移。本项目<strong>推荐并引用</strong>以下第三方免费迁移服务（均为外部网站，与本工具无隶属关系）：</p>
          <ul>
            <li><a :href="i18n.tunemyMusicUrl" target="_blank" rel="noopener">TunemyMusic（中文版）</a></li>
            <li><a href="https://spotlistr.com" target="_blank" rel="noopener">Spotlistr</a></li>
          </ul>
          <p>迁移步骤（以 TunemyMusic 为例）：</p>
          <ol>
            <li>打开上述网站，选择来源为「<strong>任意文本 / Any Text</strong>」</li>
            <li>将本工具复制的歌单文本粘贴进去</li>
            <li>选择目的地为 Apple Music / YouTube Music / Spotify 等</li>
            <li>确认并开始迁移</li>
          </ol>

          <h3>支持的平台</h3>
          <p>网易云音乐、QQ音乐、汽水音乐、酷狗音乐、酷我音乐、咪咕音乐、千千音乐、JOOX、bilibili、5Sing、Apple Music、Jamendo。</p>
          <p class="guide-note">注：酷狗概念版、波点音乐暂不支持解析。</p>

          <h3>说明</h3>
          <ul>
            <li>本工具纯本地运行，无需数据库，不收集任何数据。</li>
            <li>第三方迁移服务为外部站点，请自行判断其隐私与可用性。</li>
          </ul>
        </div>
      </aside>
    </div>

  </div>
</template>

<script setup>
import {reactive, ref, onMounted, onUnmounted} from 'vue';
import axios from 'axios';
import {isSupportedPlatform, isValidUrl} from "@/utils/utils";
import {sendErrorMessage, sendSuccessMessage} from "@/utils/tip";
import EditorTab from '@/components/EditorTab.vue';

const state = reactive({
  link: '',
  result: '',
  songsCount: 0,
  platform: '', // 来源平台标识，如 qq / kugou
  platformName: '', // 来源平台中文名，如 QQ音乐
  useDetailedSongName: false,
  songFormat: 'song-singer', // 默认为"歌名-歌手"格式
  songOrder: 'normal', // 默认为正序
  // 自定义 JSON 模板（format=custom 时使用）
  topTemplate: '{"name":"{{name}}","tracks":[{{songs}}]}',
  songTemplate: '{"name":"{{song.name}}","artist":[{{song.artists}}]}',
  // 聚合模式
  activeTab: 'single',
  urls: '',
  aggResult: '',
  aggCount: 0,
  aggDuplicates: 0,
  aggSources: [],
});

// 使用指南抽屉
const guideDrawer = ref(false);
const drawerSize = ref(typeof window !== 'undefined' && window.innerWidth < 768 ? '92%' : '460px');

const onResize = () => {
  drawerSize.value = window.innerWidth < 768 ? '92%' : '460px';
};
onMounted(() => window.addEventListener('resize', onResize));
onUnmounted(() => window.removeEventListener('resize', onResize));

const i18n = {
  title: '歌单解析与聚合',
  subtitle: '输入各平台歌单链接，一键解析并聚合为统一歌单文本',
  tabSingle: '单歌单解析',
  tabAggregate: '多歌单聚合',
  tabEditor: '文件导入编辑',
  inputPlaceholder: '输入任意歌单链接，如：http://163cn.tv/zoIxm3',
  aggInputPlaceholder: '每行输入一个歌单链接，如：\nhttp://163cn.tv/zoIxm3\nhttps://y.qq.com/n/ryqq/playlist/...',
  fetchSongList: '获取歌单',
  aggregateBtn: '聚合歌单',
  resultHint: '结果会显示在这里',
  aggResultHint: '聚合结果会显示在这里',
  songsCount: '歌曲总数',
  aggDuplicates: '已去重',
  aggSourcePlatform: '平台',
  aggSourceCount: '首数',
  aggSourceStatus: '状态',
  aggSourceOK: '成功',
  aggSourceFail: '失败',
  downloadTxt: '下载 .txt',
  guideTitle: '使用指南',
  tunemyMusicUrl: 'https://www.tunemymusic.com/zh-CN/transfer',
  copy: '复制结果',
  noContent: '没有内容可复制',
  copied: '已复制到剪贴板',
  detailedSongName: '使用未经处理的原始歌曲名',
  detailedSongNameTip: '默认不勾选此项是一种优化选择，处理后的歌曲名在迁移到其他平台时有更好的匹配率',
  emptyPlaylist: '解析失败，请检查歌单是否开放访问权限或链接是否正确。',
  songFormat: '歌曲格式',
  formatSongSinger: '歌名 - 歌手',
  formatSingerSong: '歌手 - 歌名',
  formatSongOnly: '仅歌名',
  formatCustom: '自定义 JSON',
  customTmplTop: '歌单模板（顶层）',
  customTmplSong: '单曲模板',
  customTmplTip: '占位符：歌单模板可用 {{name}}、{{count}}、{{songs}}；单曲模板可用 {{song.name}}、{{song.artist}}(歌手字符串)、{{song.artists}}(歌手数组元素)、{{song.album}}、{{song.id}}。歌曲数组需用 [{{songs}}] 包裹。默认即示例格式，可直接使用。',
  downloadJson: '下载 .json',
  songOrder: '歌曲顺序',
  orderNormal: '正序',
  orderReverse: '倒序',
};

function reset(msg) {
  sendErrorMessage(msg)
  state.result = ""
  state.songsCount = 0
  state.platform = ""
  state.platformName = ""
  state.aggResult = ""
  state.aggCount = 0
  state.aggDuplicates = 0
  state.aggSources = []
}

// 获取歌单详情（同源请求 /songlist）
const fetchLinkDetails = async () => {

  state.link = state.link.trim();

  if (!isValidUrl(state.link) || !isSupportedPlatform(state.link)) {
    reset('链接无效，支持平台：网易云/QQ/汽水/酷狗/酷我/咪咕/千千/JOOX/bilibili/5Sing/Apple/Jamendo');
    return;
  }

  const params = new URLSearchParams();
  params.append('url', state.link);
  if (state.songFormat === 'custom') {
    params.append('template_top', state.topTemplate);
    params.append('template_song', state.songTemplate);
  }

  try {
    // 构建查询参数
    let queryParams = state.useDetailedSongName ? '?detailed=true' : '?detailed=false';
    queryParams += `&format=${state.songFormat}`;
    queryParams += `&order=${state.songOrder}`;

    const resp = await axios.post('/songlist' + queryParams, params, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      },
    });

    console.log(resp.data)
    if (resp.data.code !== 1) {
      reset("请求失败，请稍后再试~");
      return;
    }

    // 自定义 JSON 模式：直接展示渲染后的 JSON 文本
    if (state.songFormat === 'custom') {
      if (!resp.data.data.json) {
        reset(i18n.emptyPlaylist);
        return;
      }
      sendSuccessMessage("歌单获取成功");
      state.result = resp.data.data.json;
      state.songsCount = resp.data.data.count || 0;
      state.platform = '';
      state.platformName = '';
      return;
    }

    // 检查是否为空歌单
    if (!resp.data.data.songs || resp.data.data.songs.length === 0 || resp.data.data.songs_count === 0) {
      reset(i18n.emptyPlaylist);
      return;
    }

    sendSuccessMessage("歌单获取成功");
    state.result = resp.data.data.songs.join('\n')
    state.songsCount = resp.data.data.songs_count;
    state.platform = resp.data.data.platform || '';
    state.platformName = resp.data.data.platform_name || '';
  } catch (err) {
    console.error(err);
    // 后端规定的错误格式 err.response.data.msg
    reset(err.response?.data?.msg || "请求失败，请稍后再试~");
  }
};

// 多歌单聚合（同源请求 /aggregate）
const aggregateLinkDetails = async () => {
  if (!state.urls.trim()) {
    reset('请输入至少一个歌单链接');
    return;
  }

  const params = new URLSearchParams();
  params.append('urls', state.urls);
  if (state.songFormat === 'custom') {
    params.append('template_top', state.topTemplate);
    params.append('template_song', state.songTemplate);
  }

  try {
    let queryParams = state.useDetailedSongName ? '?detailed=true' : '?detailed=false';
    queryParams += `&format=${state.songFormat}`;
    queryParams += `&order=${state.songOrder}`;

    const resp = await axios.post('/aggregate' + queryParams, params, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      },
    });

    console.log(resp.data)
    if (resp.data.code !== 1) {
      reset("请求失败，请稍后再试~");
      return;
    }

    const d = resp.data.data;

    // 自定义 JSON 模式：直接展示渲染后的 JSON 文本
    if (state.songFormat === 'custom') {
      if (!d.json) {
        reset(i18n.emptyPlaylist);
        return;
      }
      sendSuccessMessage("歌单聚合成功");
      state.aggResult = d.json;
      state.aggCount = d.count || 0;
      state.aggDuplicates = d.duplicates_removed || 0;
      state.aggSources = d.sources || [];
      return;
    }

    if (!d.songs || d.songs.length === 0) {
      reset(i18n.emptyPlaylist);
      return;
    }

    sendSuccessMessage("歌单聚合成功");
    state.aggResult = d.songs.join('\n');
    state.aggCount = d.songs_count;
    state.aggDuplicates = d.duplicates_removed;
    state.aggSources = d.sources || [];
  } catch (err) {
    console.error(err);
    reset(err.response?.data?.msg || "请求失败，请稍后再试~");
  }
};

// 复制结果（按当前 Tab 复制对应结果）
const copyResult = () => {
  const text = state.activeTab === 'aggregate' ? state.aggResult : state.result;
  if (!text) {
    sendErrorMessage('没有内容可复制');
    return;
  }
  const textarea = document.createElement('textarea');
  textarea.value = text;
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand('copy');
  document.body.removeChild(textarea);
  sendSuccessMessage('已复制到剪贴板');
};

// 下载聚合结果为 .txt
const downloadAgg = () => {
  if (!state.aggResult) {
    sendErrorMessage('没有内容可下载');
    return;
  }
  const blob = new Blob([state.aggResult], {type: 'text/plain;charset=utf-8'});
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'aggregate_playlist.txt';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  sendSuccessMessage('已开始下载');
};

// 下载当前结果为 .json（自定义 JSON 模式使用，单/聚合通用）
const downloadJson = (filename) => {
  const text = state.activeTab === 'aggregate' ? state.aggResult : state.result;
  if (!text) {
    sendErrorMessage('没有内容可下载');
    return;
  }
  const blob = new Blob([text], {type: 'application/json;charset=utf-8'});
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  sendSuccessMessage('已开始下载');
};

// 节流函数
const throttle = (fn, delay) => {
  let lastTime = 0;
  return function (...args) {
    const now = Date.now();
    if (now - lastTime >= delay) {
      fn.apply(this, args);
      lastTime = now;
    }
  };
};

// 使用节流包装
const throttledFetchLinkDetails = throttle(fetchLinkDetails, 1000);
const throttledAggregateLinkDetails = throttle(aggregateLinkDetails, 1000);

const debounce = (fn, delay) => {
  let timer = null;

  return function () {
    let context = this;

    let args = arguments;

    clearTimeout(timer);

    timer = setTimeout(function () {
      fn.apply(context, args);
    }, delay);
  };
};

const _ResizeObserver = window.ResizeObserver;

window.ResizeObserver = class ResizeObserver extends _ResizeObserver {
  constructor(callback) {
    callback = debounce(callback, 16);
    super(callback);
  }
};

</script>


<style>
.app-container {
  margin: 0 auto;
  max-width: 1180px;
  min-height: 100vh;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 40%);
}

.app-header {
  margin-top: 0;
  margin-bottom: 1.5em;
  padding-top: 0.75em;
}

.app-main {
  margin-top: 0.5em;
  padding: 1rem 1rem 4rem;
}

.text-center {
  text-align: center;
}

.center-col {
  display: flex;
  justify-content: center;
}

.content-col {
  max-width: 760px;
  width: 100%;
}

.compact-row {
  margin-bottom: -10px;
}

.songs-count-display {
  margin-top: -1.25em;
  color: #333;
  height: 1em;
  width: 100%;
}

.agg-dup-tag {
  margin-left: 10px;
  color: #909399;
  font-size: 13px;
}

.agg-source-wrap {
  width: 100%;
  overflow-x: auto;
}

.platform-tag {
  margin-left: 10px;
  vertical-align: middle;
  background: #f0f9eb;
  color: #529b2e;
  border: 1px solid #c2e7b0;
  border-radius: 4px;
  padding: 0 6px;
  font-size: 12px;
}

.title {
  font-size: 2em;
  margin-top: 0 !important;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.subtitle {
  margin-top: 0.35em !important;
  color: #909399;
  font-size: 0.95em;
}

.mode-tabs {
  max-width: 820px;
  margin: 0 auto;
}

/* 页签栏（原生） */
.tab-bar {
  display: flex;
  border-bottom: 2px solid #e4e7ed;
  margin-bottom: 1.2em;
}

.tab-btn {
  flex: 1;
  padding: 10px 0;
  border: none;
  background: none;
  font-size: 15px;
  color: #606266;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: color .2s, border-color .2s;
}

.tab-btn:hover {
  color: #409eff;
}

.tab-btn.active {
  color: #409eff;
  font-weight: 600;
  border-bottom-color: #409eff;
}

/* 文本输入 / 文本域（原生，外观贴近原 element-plus） */
.text-input,
.text-area {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 9px 12px;
  font-size: 14px;
  color: #303133;
  background: #fff;
  outline: none;
  transition: border-color .2s;
  font-family: inherit;
}

.text-input:focus,
.text-area:focus {
  border-color: #409eff;
}

.text-area {
  resize: vertical;
  line-height: 1.6;
}

.option-line {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  margin: 4px 0;
}

.checkbox {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  font-size: 14px;
  color: #303133;
}

.checkbox input {
  margin-right: 6px;
}

.radio-label {
  display: inline-flex;
  align-items: center;
  margin-right: 16px;
  cursor: pointer;
  font-size: 14px;
  color: #303133;
}

.radio-label input {
  margin-right: 4px;
}

.info-icon {
  margin-left: 6px;
  color: #909399;
  cursor: help;
  font-style: normal;
}

.format-label {
  margin-right: 10px;
  font-size: 14px;
  display: inline-block;
  vertical-align: middle;
}

.format-radio-group {
  display: inline-block;
  vertical-align: middle;
}

/* 自定义 JSON 模板输入框 */
.custom-tmpl-block {
  border: 1px dashed #c0c4cc;
  border-radius: 6px;
  padding: 10px 12px;
  background: #fafafa;
}

.tmpl-label {
  display: block;
  font-size: 13px;
  color: #606266;
  margin: 6px 0 4px;
  font-weight: 600;
}

.tmpl-label:first-child {
  margin-top: 0;
}

.custom-tmpl-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.6;
  margin: 8px 0 0;
}

/* 按钮（原生，分危险/主要/默认三色） */
.btn {
  display: inline-block;
  padding: 9px 20px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background: #fff;
  color: #606266;
  font-size: 14px;
  cursor: pointer;
  transition: all .2s;
}

.btn:hover {
  color: #409eff;
  border-color: #c6e2ff;
  background: #ecf5ff;
}

.btn-danger {
  background: #f56c6c;
  border-color: #f56c6c;
  color: #fff;
}

.btn-danger:hover {
  background: #f78989;
  border-color: #f78989;
  color: #fff;
}

.btn-primary {
  background: #409eff;
  border-color: #409eff;
  color: #fff;
}

.btn-primary:hover {
  background: #66b1ff;
  border-color: #66b1ff;
  color: #fff;
}

.button-center {
  margin: 12px auto;
  display: block;
  min-width: 160px;
}

/* 聚合来源表（原生 table 替代 el-table） */
.agg-source-table {
  margin-top: 0.25em;
  min-width: 280px;
  border-collapse: collapse;
  width: 100%;
  font-size: 13px;
  color: #303133;
}

.agg-source-table th,
.agg-source-table td {
  border: 1px solid #ebeef5;
  padding: 7px 10px;
  text-align: left;
}

.agg-source-table th {
  background: #f5f7fa;
  color: #909399;
  font-weight: 600;
}

.agg-source-table tbody tr:nth-child(even) {
  background: #fafafa;
}

.tag {
  display: inline-block;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 12px;
  line-height: 20px;
}

.tag-ok {
  background: #f0f9eb;
  color: #529b2e;
  border: 1px solid #c2e7b0;
}

.tag-fail {
  background: #fef0f0;
  color: #f56c6c;
  border: 1px solid #fbc4c4;
  cursor: help;
}

/* 右下角悬浮帮助按钮 */
.guide-fab {
  position: fixed;
  right: 22px;
  bottom: 26px;
  z-index: 2000;
  width: 52px !important;
  height: 52px !important;
  padding: 0 !important;
  border-radius: 50% !important;
  display: flex !important;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  color: #fff;
  background: #409eff;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(64, 158, 255, 0.45);
}

/* 使用指南抽屉（原生遮罩 + 侧边面板） */
.guide-drawer-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 3000;
  display: flex;
  justify-content: flex-end;
}

.guide-drawer {
  height: 100%;
  background: #fff;
  box-shadow: -2px 0 12px rgba(0, 0, 0, 0.15);
  overflow-y: auto;
  padding: 20px 22px 40px;
  position: relative;
  box-sizing: border-box;
}

.guide-close {
  position: absolute;
  top: 12px;
  right: 14px;
  width: 30px;
  height: 30px;
  border: none;
  background: none;
  font-size: 22px;
  line-height: 1;
  color: #909399;
  cursor: pointer;
}

.guide-close:hover {
  color: #409eff;
}

/* 使用指南抽屉内容 */
.guide-content {
  font-size: 14px;
  line-height: 1.75;
  color: #303133;
  padding-right: 24px;
}

.guide-content h3 {
  margin: 1.2em 0 0.5em;
  font-size: 1.05em;
  color: #409eff;
  border-left: 3px solid #409eff;
  padding-left: 8px;
}

.guide-content h3:first-child {
  margin-top: 0;
}

.guide-content p {
  margin: 0.4em 0;
}

.guide-content ol,
.guide-content ul {
  margin: 0.4em 0;
  padding-left: 1.4em;
}

.guide-content li {
  margin: 0.25em 0;
}

.guide-content code {
  background: #f0f2f5;
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 0.92em;
}

.guide-content pre.guide-code {
  background: #1e1e1e;
  color: #e6e6e6;
  padding: 10px 12px;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 12.5px;
  line-height: 1.55;
  white-space: pre;
}

.guide-content a {
  color: #409eff;
  text-decoration: none;
}

.guide-content a:hover {
  text-decoration: underline;
}

.guide-note {
  color: #909399;
  font-size: 0.92em;
}

/* 轻量 toast 提示 */
.toast-container {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 4000;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  pointer-events: none;
}

.toast {
  min-width: 200px;
  max-width: 80vw;
  padding: 10px 16px;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  opacity: 0;
  transform: translateY(-10px);
  transition: opacity .3s, transform .3s;
}

.toast.show {
  opacity: 1;
  transform: translateY(0);
}

.toast-error {
  background: #f56c6c;
}

.toast-success {
  background: #67c23a;
}

.toast-info {
  background: #909399;
}

/* 桌面/大屏：内容列与各控件放宽，结果框更舒展 */
@media (min-width: 1200px) {
  .content-col {
    max-width: 960px;
  }

  .mode-tabs {
    max-width: 1000px;
  }
}

/* 平板 */
@media (max-width: 1199px) {
  .content-col {
    max-width: 680px;
  }
}

/* 手机：真移动端体验 */
@media (max-width: 768px) {
  .title {
    font-size: 1.5em !important;
  }

  .subtitle {
    font-size: 0.85em;
    padding: 0 0.5em;
  }

  .app-header {
    margin-bottom: 1em;
  }

  .app-main {
    padding: 0.5rem 0.75rem 5rem;
  }

  .content-col {
    max-width: 100%;
  }

  /* 按钮全宽 */
  .button-center {
    width: 100%;
    min-width: 0;
  }

  .format-label,
  .format-radio-group {
    display: block;
    height: auto;
    line-height: 1.6;
  }

  .format-radio-group {
    margin-top: 4px;
  }

  .guide-fab {
    right: 16px;
    bottom: 18px;
    width: 46px !important;
    height: 46px !important;
    font-size: 20px;
  }
}

/* 超小屏微调 */
@media (max-width: 380px) {
  .title {
    font-size: 1.3em !important;
  }
}
</style>
