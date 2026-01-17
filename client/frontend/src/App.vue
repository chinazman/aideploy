<template>
  <div id="app">
    <!-- 主内容区 -->
    <main class="main-content">
      <!-- 网站管理视图 -->
      <div v-show="currentView === 'sites'" class="view-container">
        <div class="view-header">
          <h2>网站管理</h2>
          <div class="header-actions">
            <button @click="showConfigModal = true" class="secondary-btn">
              ⚙️ 系统设置
            </button>
            <button @click="showCreateSiteModal = true" class="primary-btn">
              + 新增网站
            </button>
          </div>
        </div>

        <!-- 网站列表 -->
        <div class="tile-list">
          <div class="tile-list-header">
            <div class="tabs">
              <div
                class="tab"
                :class="{ active: siteListTab === 'bound' }"
                @click="siteListTab = 'bound'"
              >
                已绑定网站
              </div>
              <div
                class="tab"
                :class="{ active: siteListTab === 'all' }"
                @click="siteListTab = 'all'"
              >
                全部网站
              </div>
            </div>
            <button @click="loadSites" class="icon-btn">🔄</button>
          </div>
          <div v-if="filteredSites.length === 0" class="empty-state">
            <p>{{ siteListTab === 'bound' ? '暂无已绑定网站' : '暂无网站' }}</p>
          </div>
          <div v-else class="list-items">
            <div
              v-for="site in filteredSites"
              :key="site"
              class="list-item"
            >
              <div class="item-main">
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
                <button
                  v-if="config.site_paths && config.site_paths[site]"
                  @click="showDeployModal(site)"
                  class="action-btn success"
                  title="发布"
                >
                  🚀 发布
                </button>
                <button
                  v-else
                  @click="bindDirectory(site)"
                  class="action-btn"
                  title="绑定目录"
                >
                  📁 绑定目录
                </button>
                <button
                  v-if="config.site_paths && config.site_paths[site]"
                  @click="pullFromServer(site)"
                  class="action-btn info"
                  title="从服务器覆盖本地"
                >
                  ⬇️ 下载
                </button>
                <button @click="showVersions(site)" class="action-btn" title="版本历史">📜</button>
                <button
                  v-if="siteListTab === 'bound'"
                  @click="unbindDirectory(site)"
                  class="action-btn warning"
                  title="解绑目录"
                >
                  🔓 解绑
                </button>
                <button
                  v-else
                  @click="deleteSite(site)"
                  class="action-btn danger"
                  title="删除"
                >
                  🗑️ 删除
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 系统设置对话框 -->
    <div v-if="showConfigModal" class="modal" @click.self="closeConfigModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>系统设置</h2>
          <button @click="closeConfigModal" class="icon-btn">✕</button>
        </div>
        <div class="modal-body">
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
          <div class="modal-actions">
            <button @click="closeConfigModal" class="secondary-btn">取消</button>
            <button @click="saveAndCloseConfig" class="primary-btn">保存配置</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增网站对话框 -->
    <div v-if="showCreateSiteModal" class="modal" @click.self="closeCreateSiteModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>新增网站</h2>
          <button @click="closeCreateSiteModal" class="icon-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label>网站名称</label>
            <input
              v-model="newSiteName"
              type="text"
              placeholder="请输入网站名称"
              @keyup.enter="createSite"
            />
          </div>
          <div class="input-group">
            <label>绑定目录</label>
            <div class="path-input-group">
              <input
                v-model="newSitePath"
                type="text"
                placeholder="请输入本地目录路径"
                @keyup.enter="createSite"
              />
              <button @click="selectDirectory" class="secondary-btn">选择目录</button>
            </div>
          </div>
          <div class="modal-actions">
            <button @click="closeCreateSiteModal" class="secondary-btn">取消</button>
            <button @click="createSite" :disabled="!newSiteName || !newSitePath" class="primary-btn">
              创建并绑定
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 发布对话框 -->
    <div v-if="showDeployModalFlag" class="modal" @click.self="closeDeployModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>发布网站 - {{ currentDeploySite }}</h2>
          <button @click="closeDeployModal" class="icon-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label>版本说明</label>
            <input
              v-model="deployMessage"
              type="text"
              placeholder="请输入本次发布的说明（可选）"
              @keyup.enter="executeDeploy"
            />
          </div>
          <div class="info-box">
            <p>部署目录:</p>
            <p class="info-path">{{ config.site_paths && config.site_paths[currentDeploySite] }}</p>
          </div>

          <!-- 变更信息 -->
          <div v-if="checkingChanges" class="info-box">
            <p>正在检查文件变更...</p>
          </div>
          <div v-else-if="changesResult" class="info-box">
            <p>变更信息: <strong>{{ changesResult.summary }}</strong></p>
            <div v-if="changesResult.changes.length > 0" class="changes-list">
              <div
                v-for="(change, index) in displayChanges"
                :key="index"
                class="change-item"
                :class="'change-' + change.type"
              >
                <span class="change-icon">{{ getChangeIcon(change.type) }}</span>
                <span class="change-path">{{ change.path }}</span>
                <span v-if="change.type !== 'deleted'" class="change-size">{{ formatSize(change.size) }}</span>
              </div>
              <div v-if="changesResult.changes.length > 10" class="changes-more">
                还有 {{ changesResult.changes.length - 10 }} 个文件...
              </div>
            </div>
            <div v-else class="no-changes">
              <p>没有文件变更</p>
            </div>
          </div>

          <div class="modal-actions">
            <button @click="closeDeployModal" class="secondary-btn">取消</button>
            <button
              @click="executeDeploy"
              :disabled="changesResult && !changesResult.has_changes"
              class="success-btn"
            >
              发布
            </button>
          </div>
        </div>
      </div>
    </div>

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
      newSitePath: '',
      selectedSite: '',
      deployMessage: '',
      versions: [],
      currentVersionsSite: '',
      currentDeploySite: '',
      showVersionsModal: false,
      showCreateSiteModal: false,
      showDeployModalFlag: false,
      showConfigModal: false,
      checkingChanges: false,
      changesResult: null,
      message: '',
      messageType: 'info',
      siteListTab: 'bound',
      config: {
        server_url: 'http://localhost:8080/api',
        api_key: '',
        site_paths: {}
      }
    }
  },
  computed: {
    filteredSites() {
      if (this.siteListTab === 'bound') {
        return this.sites.filter(site =>
          this.config.site_paths && this.config.site_paths[site]
        )
      }
      return this.sites
    },
    displayChanges() {
      if (!this.changesResult) return []
      // 最多显示10个变更
      return this.changesResult.changes.slice(0, 10)
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

    closeConfigModal() {
      this.showConfigModal = false
    },

    async saveAndCloseConfig() {
      await this.saveConfig()
      this.showConfigModal = false
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

      if (!this.newSitePath.trim()) {
        this.showMessage('请选择或输入绑定目录', 'error')
        return
      }

      try {
        const site = await window.go.main.App.CreateSite(this.newSiteName)
        await window.go.main.App.BindSiteDirectory(this.newSiteName, this.newSitePath.trim())

        if (this.config.site_paths) {
          this.config.site_paths[this.newSiteName] = this.newSitePath.trim()
        }

        this.showMessage('网站创建成功! 域名: ' + site.domain, 'success')
        this.newSiteName = ''
        this.newSitePath = ''
        this.closeCreateSiteModal()
        await this.loadSites()
      } catch (error) {
        this.showMessage('创建失败: ' + error, 'error')
      }
    },

    async selectDirectory() {
      try {
        const path = await window.go.main.App.SelectDirectory()
        if (path) {
          this.newSitePath = path
        }
      } catch (error) {
        this.showMessage('选择目录失败: ' + error, 'error')
      }
    },

    closeCreateSiteModal() {
      this.showCreateSiteModal = false
      this.newSiteName = ''
      this.newSitePath = ''
    },

    async bindDirectory(site) {
      try {
        const path = await window.go.main.App.SelectDirectory()
        if (!path) {
          return
        }

        await window.go.main.App.BindSiteDirectory(site, path)
        this.config.site_paths[site] = path
        this.showMessage('目录绑定成功', 'success')
      } catch (error) {
        this.showMessage('绑定目录失败: ' + error, 'error')
      }
    },

    showDeployModal(site) {
      this.currentDeploySite = site
      this.deployMessage = ''
      this.changesResult = null
      this.showDeployModalFlag = true
      this.checkChanges()
    },

    async checkChanges() {
      this.checkingChanges = true
      try {
        const result = await window.go.main.App.CheckChanges(this.currentDeploySite)
        this.changesResult = result
      } catch (error) {
        console.error('检查变更失败:', error)
        this.changesResult = null
      } finally {
        this.checkingChanges = false
      }
    },

    closeDeployModal() {
      this.showDeployModalFlag = false
      this.currentDeploySite = ''
      this.deployMessage = ''
      this.changesResult = null
    },

    async executeDeploy() {
      if (!this.currentDeploySite) {
        this.showMessage('请选择网站', 'error')
        return
      }

      if (!this.config.site_paths[this.currentDeploySite]) {
        this.showMessage('请先绑定网站目录', 'error')
        return
      }

      try {
        await window.go.main.App.DeploySite(
          this.currentDeploySite,
          this.deployMessage || '更新部署'
        )
        this.showMessage('部署成功!', 'success')
        this.closeDeployModal()
      } catch (error) {
        this.showMessage('部署失败: ' + error, 'error')
      }
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
        await this.saveConfig()
        await this.loadSites()
      } catch (error) {
        this.showMessage('删除失败: ' + error, 'error')
      }
    },

    async unbindDirectory(site) {
      if (!confirm(`确定要解绑网站 "${site}" 的目录吗？`)) {
        return
      }

      try {
        delete this.config.site_paths[site]
        await this.saveConfig()
        this.showMessage('目录解绑成功', 'success')
      } catch (error) {
        this.showMessage('解绑失败: ' + error, 'error')
      }
    },

    async pullFromServer(site) {
      const dirPath = this.config.site_paths[site]
      if (!confirm(`确定要从服务器下载网站 "${site}" 并覆盖本地目录 "${dirPath}" 吗？\n\n此操作将清空本地目录并从服务器下载最新文件。`)) {
        return
      }

      try {
        await window.go.main.App.PullSite(site)
        this.showMessage('从服务器下载成功', 'success')
      } catch (error) {
        this.showMessage('下载失败: ' + error, 'error')
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

    getChangeIcon(type) {
      const icons = {
        'added': '✨',
        'modified': '📝',
        'deleted': '🗑️'
      }
      return icons[type] || '📄'
    },

    formatSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
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
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
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

.action-btn.warning {
  background: #ff8c00;
  color: white;
}

.action-btn.warning:hover {
  background: #e67400;
}

.action-btn.success {
  background: #107c10;
  color: white;
}

.action-btn.success:hover {
  background: #0b5c0b;
}

.action-btn.info {
  background: #0078d4;
  color: white;
}

.action-btn.info:hover {
  background: #005a9e;
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
  flex-wrap: wrap;
  gap: 15px;
}

.tile-list-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 400;
}

/* Tabs */
.tabs {
  display: flex;
  gap: 0;
  border: 2px solid #e0e0e0;
  border-radius: 4px;
  overflow: hidden;
}

.tab {
  padding: 8px 20px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  background: #f5f5f5;
  color: #666;
  transition: all 0.2s;
  border-right: 1px solid #e0e0e0;
}

.tab:last-child {
  border-right: none;
}

.tab:hover {
  background: #e8e8e8;
}

.tab.active {
  background: #0078d7;
  color: white;
  border-color: #0078d7;
}

.tab.active + .tab {
  border-left: none;
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

/* Modal actions */
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

/* Path input group */
.path-input-group {
  display: flex;
  gap: 10px;
}

.path-input-group input {
  flex: 1;
}

/* Secondary button */
.secondary-btn {
  background: #666;
  color: white;
  padding: 10px 20px;
  min-height: auto;
}

.secondary-btn:hover {
  background: #555;
}

/* Info box */
.info-box {
  background: #f0f0f0;
  padding: 15px;
  border-radius: 4px;
  margin: 15px 0;
}

.info-box p {
  margin: 0 0 5px 0;
  font-size: 13px;
  color: #666;
}

.info-box .info-path {
  margin: 5px 0 0 0;
  font-size: 14px;
  color: #0078d7;
  font-weight: 500;
  word-break: break-all;
}

/* Changes list */
.changes-list {
  margin-top: 10px;
  max-height: 200px;
  overflow-y: auto;
  background: white;
  border-radius: 4px;
  padding: 8px;
}

.change-item {
  display: flex;
  align-items: center;
  padding: 6px 8px;
  gap: 8px;
  font-size: 13px;
  border-bottom: 1px solid #f0f0f0;
}

.change-item:last-child {
  border-bottom: none;
}

.change-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.change-path {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #333;
}

.change-size {
  flex-shrink: 0;
  font-size: 12px;
  color: #999;
}

.change-added {
  background: #e6fffa;
}

.change-modified {
  background: #fffaf0;
}

.change-deleted {
  background: #fff5f5;
  opacity: 0.7;
}

.changes-more {
  padding: 8px;
  text-align: center;
  font-size: 12px;
  color: #999;
  font-style: italic;
}

.no-changes {
  text-align: center;
  padding: 20px;
  color: #999;
  font-size: 14px;
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
