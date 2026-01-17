<template>
  <div id="app">
    <!-- 侧边导航栏 -->
    <nav class="sidebar">
      <div class="sidebar-header">
        <h1>AI Deploy</h1>
      </div>
      <div class="nav-items">
        <div
          class="nav-item"
          :class="{ active: currentView === 'sites' }"
          @click="currentView = 'sites'"
        >
          <div class="nav-icon">🌐</div>
          <div class="nav-text">网站管理</div>
        </div>
        <div
          class="nav-item"
          :class="{ active: currentView === 'config' }"
          @click="currentView = 'config'"
        >
          <div class="nav-icon">⚙️</div>
          <div class="nav-text">系统配置</div>
        </div>
      </div>
    </nav>

    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 网站管理视图 -->
      <div v-show="currentView === 'sites'" class="view-container">
        <div class="view-header">
          <h2>网站管理</h2>
        </div>

        <div class="content-grid">
          <!-- 创建网站卡片 -->
          <div class="tile">
            <div class="tile-header">
              <h3>创建新网站</h3>
            </div>
            <div class="tile-body">
              <div class="input-group">
                <input
                  v-model="newSiteName"
                  type="text"
                  placeholder="网站名称"
                  @keyup.enter="createSite"
                />
              </div>
              <button @click="createSite" :disabled="!newSiteName" class="primary-btn">
                创建
              </button>
            </div>
          </div>

          <!-- 部署网站卡片 -->
          <div class="tile">
            <div class="tile-header">
              <h3>部署网站</h3>
            </div>
            <div class="tile-body">
              <div class="input-group">
                <select v-model="selectedSite">
                  <option value="">选择网站</option>
                  <option v-for="site in sites" :key="site" :value="site">
                    {{ site }}
                  </option>
                </select>
              </div>
              <div class="input-group">
                <input
                  type="text"
                  v-model="deployMessage"
                  placeholder="版本说明 (可选)"
                />
              </div>
              <button
                @click="deploySite"
                :disabled="!selectedSite"
                class="success-btn"
              >
                部署
              </button>
              <p class="hint-text">将使用绑定的目录进行部署</p>
            </div>
          </div>
        </div>

        <!-- 网站列表 -->
        <div class="tile-list">
          <div class="tile-list-header">
            <h3>网站列表</h3>
            <button @click="loadSites" class="icon-btn">🔄</button>
          </div>
          <div v-if="sites.length === 0" class="empty-state">
            <p>暂无网站</p>
          </div>
          <div v-else class="list-items">
            <div
              v-for="site in sites"
              :key="site"
              class="list-item"
              :class="{ active: selectedSite === site }"
            >
              <div class="item-main" @click="selectSite(site)">
                <div class="item-icon">🌐</div>
                <div class="item-content">
                  <div class="item-title">{{ site }}</div>
                  <div class="item-subtitle" v-if="config.site_paths && config.site_paths[site]">
                    {{ config.site_paths[site] }}
                  </div>
                  <div class="item-subtitle" v-else>
                    未绑定目录
                  </div>
                </div>
              </div>
              <div class="item-actions">
                <button @click="bindDirectory(site)" class="action-btn" title="绑定目录">📁</button>
                <button @click="showVersions(site)" class="action-btn" title="版本历史">📜</button>
                <button @click="deleteSite(site)" class="action-btn danger" title="删除">🗑️</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 配置视图 -->
      <div v-show="currentView === 'config'" class="view-container">
        <div class="view-header">
          <h2>系统配置</h2>
        </div>

        <div class="tile-list">
          <div class="tile">
            <div class="tile-header">
              <h3>服务器配置</h3>
            </div>
            <div class="tile-body">
              <div class="input-group">
                <label>服务器地址</label>
                <input
                  v-model="config.server_url"
                  type="text"
                  placeholder="http://localhost:8080/api"
                />
              </div>
              <div class="input-group">
                <label>API 密钥</label>
                <input
                  v-model="config.api_key"
                  type="password"
                  placeholder="可选"
                />
              </div>
              <button @click="saveConfig" class="primary-btn">保存配置</button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 版本历史对话框 -->
    <div v-if="showVersionsModal" class="modal" @click.self="closeVersionsModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>版本历史 - {{ currentVersionsSite }}</h2>
          <button @click="closeVersionsModal" class="icon-btn">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="versions.length === 0" class="empty-state">
            <p>暂无版本记录</p>
          </div>
          <div v-else class="versions-list">
            <div
              v-for="(version, index) in versions"
              :key="version.hash"
              class="version-item"
            >
              <div class="version-header">
                <span class="version-hash">{{ version.hash.substring(0, 7) }}</span>
                <span class="version-date">{{ formatDate(version.date) }}</span>
              </div>
              <div class="version-message">{{ version.message }}</div>
              <div class="version-author">👤 {{ version.author }}</div>
              <button @click="rollbackTo(version.hash)" class="warning-btn">
                ↩️ 回滚
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 消息提示 -->
    <div v-if="message" class="toast" :class="messageType">
      {{ message }}
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      currentView: 'sites',
      sites: [],
      newSiteName: '',
      selectedSite: '',
      deployMessage: '',
      versions: [],
      currentVersionsSite: '',
      showVersionsModal: false,
      message: '',
      messageType: 'info',
      config: {
        server_url: 'http://localhost:8080/api',
        api_key: '',
        site_paths: {}
      }
    }
  },
  mounted() {
    this.loadConfig()
    this.loadSites()
  },
  methods: {
    async loadConfig() {
      try {
        const config = await window.go.main.App.GetConfig()
        this.config = config
      } catch (error) {
        console.error('加载配置失败:', error)
      }
    },

    async saveConfig() {
      try {
        await window.go.main.App.SaveConfig(this.config)
        this.showMessage('配置保存成功', 'success')
      } catch (error) {
        this.showMessage('保存配置失败: ' + error, 'error')
      }
    },

    async loadSites() {
      try {
        const sites = await window.go.main.App.ListSites()
        this.sites = sites
      } catch (error) {
        this.showMessage('加载网站列表失败: ' + error, 'error')
      }
    },

    async createSite() {
      if (!this.newSiteName.trim()) {
        this.showMessage('请输入网站名称', 'error')
        return
      }

      try {
        const site = await window.go.main.App.CreateSite(this.newSiteName)
        this.showMessage('网站创建成功! 域名: ' + site.domain, 'success')
        this.newSiteName = ''
        await this.loadSites()
      } catch (error) {
        this.showMessage('创建失败: ' + error, 'error')
      }
    },

    async bindDirectory(site) {
      const path = prompt('请输入本地目录路径:', this.config.site_paths[site] || '')
      if (path === null) return

      if (!path.trim()) {
        this.showMessage('目录路径不能为空', 'error')
        return
      }

      try {
        await window.go.main.App.BindSiteDirectory(site, path.trim())
        this.config.site_paths[site] = path.trim()
        this.showMessage('目录绑定成功', 'success')
      } catch (error) {
        this.showMessage('绑定目录失败: ' + error, 'error')
      }
    },

    async deploySite() {
      if (!this.selectedSite) {
        this.showMessage('请选择网站', 'error')
        return
      }

      if (!this.config.site_paths[this.selectedSite]) {
        this.showMessage('请先绑定网站目录', 'error')
        return
      }

      try {
        await window.go.main.App.DeploySite(
          this.selectedSite,
          this.deployMessage || '更新部署'
        )
        this.showMessage('部署成功!', 'success')
        this.deployMessage = ''
      } catch (error) {
        this.showMessage('部署失败: ' + error, 'error')
      }
    },

    selectSite(site) {
      this.selectedSite = site
    },

    async deleteSite(site) {
      if (!confirm(`确定要删除网站 "${site}" 吗？此操作不可恢复！`)) {
        return
      }

      try {
        await window.go.main.App.DeleteSite(site)
        this.showMessage('网站删除成功', 'success')
        if (this.selectedSite === site) {
          this.selectedSite = ''
        }
        delete this.config.site_paths[site]
        await this.loadSites()
      } catch (error) {
        this.showMessage('删除失败: ' + error, 'error')
      }
    },

    async showVersions(site) {
      this.currentVersionsSite = site
      try {
        const versions = await window.go.main.App.GetVersions(site)
        this.versions = versions
        this.showVersionsModal = true
      } catch (error) {
        this.showMessage('获取版本失败: ' + error, 'error')
      }
    },

    closeVersionsModal() {
      this.showVersionsModal = false
      this.currentVersionsSite = ''
      this.versions = []
    },

    async rollbackTo(hash) {
      const message = prompt('请输入回滚说明:', '回滚到版本 ' + hash.substring(0, 7))
      if (message === null) return

      try {
        await window.go.main.App.Rollback(
          this.currentVersionsSite,
          hash,
          message
        )
        this.showMessage('回滚成功!', 'success')
        this.closeVersionsModal()
      } catch (error) {
        this.showMessage('回滚失败: ' + error, 'error')
      }
    },

    formatDate(dateStr) {
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN')
    },

    showMessage(msg, type = 'info') {
      this.message = msg
      this.messageType = type
      setTimeout(() => {
        this.message = ''
      }, 3000)
    }
  }
}
</script>

<style scoped>
/* Metro 风格全局样式 */
#app {
  display: flex;
  min-height: 100vh;
  background: #f0f0f0;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}

/* 侧边导航栏 */
.sidebar {
  width: 240px;
  background: #1e1e1e;
  color: white;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0,0,0,0.1);
}

.sidebar-header {
  padding: 40px 20px 30px;
  border-bottom: 1px solid #333;
}

.sidebar-header h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 300;
  letter-spacing: 2px;
}

.nav-items {
  flex: 1;
  padding: 20px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 15px 20px;
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.nav-item:hover {
  background: #2d2d2d;
}

.nav-item.active {
  background: #0078d7;
  border-left-color: #fff;
}

.nav-icon {
  font-size: 24px;
  margin-right: 12px;
}

.nav-text {
  font-size: 14px;
  font-weight: 500;
}

/* 主内容区 */
.main-content {
  flex: 1;
  overflow-y: auto;
  background: #f0f0f0;
}

.view-container {
  padding: 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.view-header {
  margin-bottom: 30px;
}

.view-header h2 {
  margin: 0;
  font-size: 32px;
  font-weight: 300;
  color: #1e1e1e;
}

/* Metro Tile 布局 */
.content-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.tile {
  background: white;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  transition: all 0.2s;
  cursor: default;
}

.tile:hover {
  box-shadow: 0 4px 8px rgba(0,0,0,0.15);
  transform: translateY(-2px);
}

.tile-header {
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 2px solid #0078d7;
}

.tile-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 400;
  color: #1e1e1e;
}

.tile-body {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

/* 输入组 */
.input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input-group label {
  font-size: 13px;
  font-weight: 600;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.input-group input,
.input-group select {
  padding: 10px 12px;
  border: 2px solid #e0e0e0;
  background: white;
  font-size: 14px;
  transition: all 0.2s;
  box-sizing: border-box;
}

.input-group input:focus,
.input-group select:focus {
  outline: none;
  border-color: #0078d7;
}

.hint-text {
  margin: 0;
  font-size: 12px;
  color: #999;
  font-style: italic;
}

/* Metro 风格按钮 */
button {
  padding: 12px 24px;
  border: none;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  min-height: 40px;
}

button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

button:active:not(:disabled) {
  transform: translateY(0);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary-btn {
  background: #0078d7;
  color: white;
}

.primary-btn:hover:not(:disabled) {
  background: #005a9e;
}

.success-btn {
  background: #107c10;
  color: white;
}

.success-btn:hover:not(:disabled) {
  background: #0b5c0b;
}

.warning-btn {
  background: #d83b01;
  color: white;
}

.warning-btn:hover:not(:disabled) {
  background: #a52c00;
}

.icon-btn {
  background: transparent;
  color: #666;
  padding: 8px;
  min-height: auto;
  font-size: 18px;
}

.icon-btn:hover {
  background: #e0e0e0;
}

.action-btn {
  background: #e0e0e0;
  color: #333;
  padding: 8px 12px;
  min-height: auto;
  margin-left: 8px;
}

.action-btn:hover {
  background: #d0d0d0;
}

.action-btn.danger {
  background: #d13438;
  color: white;
}

.action-btn.danger:hover {
  background: #a52c00;
}

/* 列表样式 */
.tile-list {
  background: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.tile-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
}

.tile-list-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 400;
}

.list-items {
  display: flex;
  flex-direction: column;
}

.list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  transition: all 0.2s;
}

.list-item:hover {
  background: #f8f8f8;
}

.list-item.active {
  background: #e8f4ff;
  border-left: 3px solid #0078d7;
}

.item-main {
  display: flex;
  align-items: center;
  flex: 1;
  cursor: pointer;
}

.item-icon {
  font-size: 28px;
  margin-right: 16px;
}

.item-content {
  flex: 1;
}

.item-title {
  font-size: 16px;
  font-weight: 500;
  color: #1e1e1e;
  margin-bottom: 4px;
}

.item-subtitle {
  font-size: 13px;
  color: #666;
}

.item-actions {
  display: flex;
  align-items: center;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-state p {
  margin: 0;
  font-size: 16px;
}

/* 模态框 */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  max-width: 700px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 8px 16px rgba(0,0,0,0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 2px solid #0078d7;
}

.modal-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 400;
}

.modal-body {
  padding: 20px;
}

.versions-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.version-item {
  padding: 16px;
  border: 1px solid #e0e0e0;
  background: #fafafa;
}

.version-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.version-hash {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 600;
  color: #0078d7;
  font-size: 14px;
}

.version-date {
  color: #666;
  font-size: 13px;
}

.version-message {
  margin: 8px 0;
  color: #333;
  font-size: 14px;
}

.version-author {
  color: #666;
  font-size: 13px;
  margin-bottom: 12px;
}

/* Toast 提示 */
.toast {
  position: fixed;
  bottom: 30px;
  right: 30px;
  padding: 16px 24px;
  color: white;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  animation: slideIn 0.3s ease;
  z-index: 2000;
  min-width: 300px;
}

.toast.success {
  background: #107c10;
}

.toast.error {
  background: #d13438;
}

.toast.info {
  background: #0078d7;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

/* 响应式设计 */
@media (max-width: 768px) {
  #app {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
  }

  .nav-items {
    display: flex;
    overflow-x: auto;
  }

  .nav-item {
    flex: 1;
    justify-content: center;
    border-left: none;
    border-bottom: 3px solid transparent;
  }

  .nav-item.active {
    border-left: none;
    border-bottom-color: #fff;
  }

  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
