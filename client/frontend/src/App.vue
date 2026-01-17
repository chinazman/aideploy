<template>
  <div id="app">
    <div class="container">
      <header>
        <h1>🚀 AI原型部署工具</h1>
        <p class="subtitle">快速发布您的HTML原型</p>
      </header>

      <div class="main-content">
        <!-- 左侧面板 - 操作区 -->
        <div class="left-panel">
          <div class="card">
            <h2>创建新网站</h2>
            <div class="form-group">
              <label>网站名称</label>
              <input
                v-model="newSiteName"
                type="text"
                placeholder="例如: my-prototype"
                @keyup.enter="createSite"
              />
            </div>
            <button @click="createSite" :disabled="!newSiteName">
              创建网站
            </button>
          </div>

          <div class="card">
            <h2>部署网站</h2>
            <div class="form-group">
              <label>选择网站</label>
              <select v-model="selectedSite">
                <option value="">-- 请选择 --</option>
                <option v-for="site in sites" :key="site" :value="site">
                  {{ site }}
                </option>
              </select>
            </div>
            <div class="form-group">
              <label>选择文件</label>
              <input
                type="file"
                ref="fileInput"
                accept=".html,.htm"
                @change="handleFileSelect"
              />
            </div>
            <div class="form-group">
              <label>版本说明</label>
              <input
                v-model="deployMessage"
                type="text"
                placeholder="例如: 更新首页设计"
              />
            </div>
            <button
              @click="deploySite"
              :disabled="!selectedSite || !selectedFile"
            >
              部署
            </button>
          </div>
        </div>

        <!-- 右侧面板 - 网站列表 -->
        <div class="right-panel">
          <div class="card">
            <div class="card-header">
              <h2>网站列表</h2>
              <button @click="loadSites" class="refresh-btn">🔄 刷新</button>
            </div>
            <div v-if="sites.length === 0" class="empty-state">
              <p>暂无网站</p>
              <p class="hint">创建一个网站开始部署</p>
            </div>
            <div v-else class="site-list">
              <div
                v-for="site in sites"
                :key="site"
                class="site-item"
                :class="{ active: selectedSite === site }"
                @click="selectSite(site)"
              >
                <div class="site-info">
                  <h3>{{ site }}</h3>
                  <button @click.stop="showVersions(site)" class="view-versions-btn">
                    📜 查看版本
                  </button>
                  <button @click.stop="deleteSite(site)" class="delete-btn">
                    🗑️ 删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 版本历史对话框 -->
      <div v-if="showVersionsModal" class="modal" @click.self="closeVersionsModal">
        <div class="modal-content">
          <div class="modal-header">
            <h2>版本历史 - {{ currentVersionsSite }}</h2>
            <button @click="closeVersionsModal" class="close-btn">✕</button>
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
                <button @click="rollbackTo(version.hash)" class="rollback-btn">
                  ↩️ 回滚到此版本
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 消息提示 -->
      <div v-if="message" class="message" :class="messageType">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      sites: [],
      newSiteName: '',
      selectedSite: '',
      selectedFile: null,
      deployMessage: '',
      versions: [],
      currentVersionsSite: '',
      showVersionsModal: false,
      message: '',
      messageType: 'info'
    }
  },
  mounted() {
    this.loadSites()
  },
  methods: {
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

    handleFileSelect(event) {
      this.selectedFile = event.target.files[0]
    },

    async deploySite() {
      if (!this.selectedSite || !this.selectedFile) {
        this.showMessage('请选择网站和文件', 'error')
        return
      }

      try {
        await window.go.main.App.DeploySite(
          this.selectedSite,
          this.selectedFile.path,
          this.deployMessage || '更新部署'
        )
        this.showMessage('部署成功!', 'success')
        this.deployMessage = ''
        this.selectedFile = null
        this.$refs.fileInput.value = ''
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
#app {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.container {
  max-width: 1400px;
  margin: 0 auto;
}

header {
  text-align: center;
  color: white;
  margin-bottom: 30px;
}

header h1 {
  font-size: 2.5em;
  margin: 0;
  text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
}

.subtitle {
  font-size: 1.2em;
  opacity: 0.9;
  margin-top: 10px;
}

.main-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.left-panel, .right-panel {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.card h2 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #333;
  font-size: 1.5em;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.refresh-btn {
  background: #667eea;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
}

.refresh-btn:hover {
  background: #5568d3;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #555;
  font-weight: 500;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px;
  border: 2px solid #e0e0e0;
  border-radius: 6px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
}

button {
  background: #667eea;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  width: 100%;
  transition: background 0.2s;
}

button:hover:not(:disabled) {
  background: #5568d3;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.site-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.site-item {
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.site-item:hover {
  border-color: #667eea;
  background: #f8f9ff;
}

.site-item.active {
  border-color: #667eea;
  background: #f0f2ff;
}

.site-info h3 {
  margin: 0 0 12px 0;
  color: #333;
}

.site-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.view-versions-btn {
  background: #48bb78;
  padding: 6px 12px;
  font-size: 14px;
  width: auto;
}

.view-versions-btn:hover {
  background: #38a169;
}

.delete-btn {
  background: #f56565;
  padding: 6px 12px;
  font-size: 14px;
  width: auto;
}

.delete-btn:hover {
  background: #e53e3e;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #999;
}

.empty-state p {
  margin: 8px 0;
}

.hint {
  font-size: 0.9em;
}

/* 模态框 */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 12px;
  padding: 24px;
  max-width: 700px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h2 {
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  width: auto;
  padding: 0;
  color: #999;
}

.close-btn:hover {
  color: #333;
}

.versions-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.version-item {
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  padding: 16px;
  position: relative;
}

.version-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.version-hash {
  font-family: monospace;
  font-weight: bold;
  color: #667eea;
}

.version-date {
  color: #999;
  font-size: 0.9em;
}

.version-message {
  margin: 8px 0;
  color: #333;
}

.version-author {
  color: #666;
  font-size: 0.9em;
  margin-bottom: 12px;
}

.rollback-btn {
  background: #ed8936;
  font-size: 14px;
  padding: 8px 16px;
  width: auto;
}

.rollback-btn:hover {
  background: #dd6b20;
}

/* 消息提示 */
.message {
  position: fixed;
  bottom: 20px;
  right: 20px;
  padding: 16px 24px;
  border-radius: 8px;
  color: white;
  font-weight: 500;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
  animation: slideIn 0.3s ease;
  max-width: 400px;
  z-index: 2000;
}

.message.success {
  background: #48bb78;
}

.message.error {
  background: #f56565;
}

.message.info {
  background: #4299e1;
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

@media (max-width: 768px) {
  .main-content {
    grid-template-columns: 1fr;
  }
}
</style>
