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
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              系统设置
            </button>
            <button @click="showCreateSiteModal = true" class="primary-btn">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
              新增网站
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
            <button @click="loadSites" class="icon-btn" title="刷新列表">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
              </svg>
            </button>
          </div>
          <div v-if="filteredSites.length === 0" class="empty-state">
            <p>{{ siteListTab === 'bound' ? '暂无已绑定网站' : '暂无网站' }}</p>
          </div>
          <div v-else class="list-items">
            <div
              v-for="site in filteredSites"
              :key="site.name"
              class="list-item"
            >
              <div class="item-main" @click="openSiteInBrowser(site.url)">
                <div class="item-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S12 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S7.5 3 12 3m0-18a9 9 0 018.716 6.747M12 3a9 9 0 00-8.716 6.747M12 3c2.485 0 4.5 4.03 4.5 9s-2.015 9-4.5 9m0-18c-2.485 0-4.5 4.03-4.5 9s2.015 9 4.5 9" />
                  </svg>
                </div>
                <div class="item-content">
                  <div class="item-title">{{ site.name }}</div>
                  <div class="item-subtitle" v-if="config.site_paths && config.site_paths[site.name]">
                    {{ config.site_paths[site.name] }}
                  </div>
                  <div class="item-subtitle" v-else>
                    未绑定目录
                  </div>
                </div>
              </div>
              <div class="item-actions">
                <button
                  v-if="config.site_paths && config.site_paths[site.name]"
                  @click="openDirectory(config.site_paths[site.name])"
                  class="action-btn"
                  title="打开目录"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
                  </svg>
                </button>
                <button
                  v-if="config.site_paths && config.site_paths[site.name]"
                  @click="showDeployModal(site.name)"
                  class="action-btn success"
                  title="发布"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.581-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z" />
                  </svg>
                </button>
                <button
                  v-else
                  @click="bindDirectory(site.name)"
                  class="action-btn"
                  title="绑定目录"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" />
                  </svg>
                </button>
                <button
                  v-if="config.site_paths && config.site_paths[site.name]"
                  @click="pullFromServer(site.name)"
                  class="action-btn info"
                  title="从服务器覆盖本地"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
                  </svg>
                </button>
                <button @click="showVersions(site.name)" class="action-btn" title="版本历史">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </button>
                <button
                  v-if="siteListTab === 'bound'"
                  @click="unbindDirectory(site.name)"
                  class="action-btn warning"
                  title="解绑目录"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 10.5V6.75a4.5 4.5 0 119 9v3.75M3.75 21.75h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H3.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
                  </svg>
                </button>
                <button
                  v-else
                  @click="deleteSite(site.name)"
                  class="action-btn danger"
                  title="删除"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                  </svg>
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
          <button @click="closeConfigModal" class="icon-btn">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
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
            <label>用户名</label>
            <input
              v-model="config.username"
              type="text"
              placeholder="请输入用户名"
            />
          </div>
          <div class="input-group">
            <label>密码</label>
            <input
              v-model="config.password"
              type="password"
              placeholder="请输入密码"
            />
          </div>
          <div class="modal-actions">
            <button @click="closeConfigModal" class="secondary-btn">取消</button>
            <button @click="saveAndCloseConfig" class="primary-btn">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width: 18px; height: 18px;">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
              </svg>
              保存配置
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增网站对话框 -->
    <div v-if="showCreateSiteModal" class="modal" @click.self="closeCreateSiteModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>新增网站</h2>
          <button @click="closeCreateSiteModal" class="icon-btn">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
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
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width: 18px; height: 18px;">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
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
          <button @click="closeDeployModal" class="icon-btn">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
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
                <span class="change-icon">
                  <svg v-if="change.type === 'added'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                  </svg>
                  <svg v-else-if="change.type === 'modified'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                  </svg>
                  <svg v-else-if="change.type === 'deleted'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                  </svg>
                  <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                  </svg>
                </span>
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
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" style="width: 18px; height: 18px;">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.581-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z" />
              </svg>
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
          <button @click="closeVersionsModal" class="icon-btn">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
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
              <div class="version-author">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                </svg>
                {{ version.author }}
              </div>
              <button @click="rollbackTo(version.hash)" class="warning-btn">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 15L3 9m0 0l6-6M3 9h12a6 6 0 010 12h-3" />
                </svg>
                回滚
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 消息提示 -->
    <div v-if="message" class="toast" :class="messageType">
      <div class="toast-icon">
        <svg v-if="messageType === 'success'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <svg v-else-if="messageType === 'error'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
        </svg>
      </div>
      <div class="toast-content">{{ message }}</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      currentView: 'sites',
      sites: [], // 改为存储网站对象数组 {name, url}
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
        username: '',
        password: '',
        site_paths: {}
      }
    }
  },
  computed: {
    filteredSites() {
      if (this.siteListTab === 'bound') {
        return this.sites.filter(site =>
          this.config.site_paths && this.config.site_paths[site.name]
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

    async openDirectory(path) {
      try {
        await window.go.main.App.OpenDirectory(path)
      } catch (error) {
        this.showMessage('打开目录失败: ' + error, 'error')
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
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return date.toLocaleString('zh-CN')
    },

    formatSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },

    openSiteInBrowser(url) {
      // 在浏览器中打开网站
      window.open(url, '_blank')
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
/* Modern Tech 风格全局样式 */
#app {
  display: flex;
  min-height: 100vh;
  background: #0f172a; /* Slate 900 */
  background-image:
    radial-gradient(at 0% 0%, rgba(56, 189, 248, 0.15) 0px, transparent 50%),
    radial-gradient(at 100% 0%, rgba(139, 92, 246, 0.15) 0px, transparent 50%);
  font-family: 'Segoe UI', 'Inter', -apple-system, BlinkMacSystemFont, Roboto, Helvetica, Arial, sans-serif;
  color: #e2e8f0; /* Slate 200 */
}

/* 主内容区 */
.main-content {
  flex: 1;
  overflow-y: auto;
  background: transparent;
}

.view-container {
  padding: 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.view-header {
  margin-bottom: 40px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.view-header h2 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
  background: linear-gradient(135deg, #38bdf8 0%, #818cf8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.5px;
}

/* 列表样式 */
.tile-list {
  background: rgba(30, 41, 59, 0.4); /* Slate 800 with opacity */
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.tile-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-wrap: wrap;
  gap: 15px;
}

/* Tabs */
.tabs {
  display: flex;
  background: rgba(15, 23, 42, 0.5);
  padding: 4px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.tab {
  padding: 8px 20px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #94a3b8; /* Slate 400 */
  transition: all 0.3s ease;
  border-radius: 6px;
  border: none;
}

.tab:hover {
  color: #e2e8f0;
}

.tab.active {
  background: rgba(56, 189, 248, 0.1);
  color: #38bdf8; /* Sky 400 */
}

.list-items {
  display: flex;
  flex-direction: column;
}

.list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  transition: all 0.2s ease;
}

.list-item:last-child {
  border-bottom: none;
}

.list-item:hover {
  background: rgba(255, 255, 255, 0.03);
}

.item-main {
  display: flex;
  align-items: center;
  flex: 1;
  cursor: pointer;
  border-radius: 8px;
  padding: 8px;
  margin-right: 8px;
  transition: background 0.2s ease;
}

.item-main:hover {
  background: rgba(56, 189, 248, 0.1);
}

.item-icon {
  width: 40px;
  height: 40px;
  margin-right: 20px;
  color: #38bdf8;
  flex-shrink: 0;
  filter: drop-shadow(0 0 8px rgba(56, 189, 248, 0.3));
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-icon svg {
  width: 32px;
  height: 32px;
}

.item-content {
  flex: 1;
}

.item-title {
  font-size: 16px;
  font-weight: 600;
  color: #f1f5f9; /* Slate 100 */
  margin-bottom: 6px;
}

.item-subtitle {
  font-size: 13px;
  color: #64748b; /* Slate 500 */
  font-family: monospace;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
  color: #64748b;
}

.empty-state p {
  margin: 0;
  font-size: 16px;
}

/* 按钮通用样式 */
button {
  padding: 10px 20px;
  border: none;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  line-height: 1.5;
}

button svg {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

button:hover:not(:disabled) {
  transform: translateY(-1px);
}

button:active:not(:disabled) {
  transform: translateY(0);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  filter: grayscale(1);
}

.primary-btn {
  background: linear-gradient(135deg, #38bdf8 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(56, 189, 248, 0.3);
}

.primary-btn:hover:not(:disabled) {
  box-shadow: 0 6px 16px rgba(56, 189, 248, 0.4);
}

.secondary-btn {
  background: rgba(255, 255, 255, 0.05);
  color: #e2e8f0;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.secondary-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

.success-btn {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.warning-btn {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.warning-btn:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
}

.icon-btn {
  background: transparent;
  color: #94a3b8;
  padding: 8px;
  min-height: auto;
  font-size: 18px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.icon-btn svg {
  width: 24px;
  height: 24px;
}

.icon-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #f1f5f9;
}

.action-btn {
  background: rgba(30, 41, 59, 0.6);
  color: #cbd5e1;
  padding: 8px;
  width: 36px;
  height: 36px;
  font-size: 13px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.action-btn svg {
  width: 20px;
  height: 20px;
}

.action-btn:hover {
  background: rgba(51, 65, 85, 0.8);
  color: white;
  border-color: rgba(255, 255, 255, 0.15);
}

.action-btn.success {
  color: #34d399;
  border-color: rgba(52, 211, 153, 0.3);
  background: rgba(52, 211, 153, 0.1);
}

.action-btn.success:hover {
  background: rgba(52, 211, 153, 0.2);
}

.action-btn.info {
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.3);
  background: rgba(56, 189, 248, 0.1);
}

.action-btn.info:hover {
  background: rgba(56, 189, 248, 0.2);
}

.action-btn.warning {
  color: #fbbf24;
  border-color: rgba(251, 191, 36, 0.3);
  background: rgba(251, 191, 36, 0.1);
}

.action-btn.warning:hover {
  background: rgba(251, 191, 36, 0.2);
}

.action-btn.danger {
  color: #f87171;
  border-color: rgba(248, 113, 113, 0.3);
  background: rgba(248, 113, 113, 0.1);
}

.action-btn.danger:hover {
  background: rgba(248, 113, 113, 0.2);
}

/* 模态框 */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

.modal-content {
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  max-width: 700px;
  width: 90%;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(30, 41, 59, 0.5);
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #f1f5f9;
}

.modal-body {
  padding: 30px;
}

/* 输入组 */
.input-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
}

.input-group label {
  font-size: 13px;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.input-group input,
.input-group select {
  padding: 12px 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: #0f172a;
  color: #f1f5f9;
  font-size: 14px;
  border-radius: 8px;
  transition: all 0.2s;
  box-sizing: border-box;
}

.input-group input:focus,
.input-group select:focus {
  outline: none;
  border-color: #38bdf8;
  box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.2);
}

.path-input-group {
  display: flex;
  gap: 10px;
}

.path-input-group input {
  flex: 1;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

/* Info box */
.info-box {
  background: rgba(56, 189, 248, 0.05);
  padding: 20px;
  border-radius: 12px;
  border: 1px solid rgba(56, 189, 248, 0.1);
  margin: 20px 0;
}

.info-box p {
  margin: 0 0 8px 0;
  font-size: 13px;
  color: #94a3b8;
}

.info-box .info-path {
  margin: 0;
  font-size: 14px;
  color: #38bdf8;
  font-family: monospace;
  word-break: break-all;
}

/* Changes list */
.changes-list {
  margin-top: 15px;
  max-height: 250px;
  overflow-y: auto;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 4px;
}

.change-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  gap: 10px;
  font-size: 13px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.change-item:last-child {
  border-bottom: none;
}

.change-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
}

.change-icon svg {
  width: 16px;
  height: 16px;
}

.change-path {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #cbd5e1;
  font-family: monospace;
}

.change-added {
  background: rgba(16, 185, 129, 0.1);
}

.change-added .change-icon {
  color: #34d399;
}

.change-modified {
  background: rgba(245, 158, 11, 0.1);
}

.change-modified .change-icon {
  color: #fbbf24;
}

.change-deleted {
  background: rgba(239, 68, 68, 0.1);
  opacity: 0.8;
}

.change-deleted .change-icon {
  color: #f87171;
}

.changes-more {
  padding: 12px;
  text-align: center;
  font-size: 12px;
  color: #64748b;
}

/* Versions List */
.versions-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.version-item {
  padding: 20px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(30, 41, 59, 0.4);
  border-radius: 12px;
}

.version-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.version-hash {
  font-family: monospace;
  font-weight: 600;
  color: #38bdf8;
  font-size: 14px;
  background: rgba(56, 189, 248, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.version-date {
  color: #64748b;
  font-size: 13px;
}

.version-message {
  margin: 12px 0;
  color: #e2e8f0;
  font-size: 15px;
}

.version-author {
  color: #94a3b8;
  font-size: 13px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.version-author svg {
  width: 16px;
  height: 16px;
}

/* Toast 提示 */
.toast {
  position: fixed;
  bottom: 40px;
  right: 40px;
  padding: 16px 24px;
  color: white;
  font-weight: 500;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.5);
  animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 2000;
  min-width: 300px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  backdrop-filter: blur(8px);
}

.toast-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
}

.toast-icon svg {
  width: 24px;
  height: 24px;
}

.toast-content {
  flex: 1;
}

.toast.success {
  background: rgba(16, 185, 129, 0.9);
  border: 1px solid rgba(16, 185, 129, 0.5);
}

.toast.error {
  background: rgba(239, 68, 68, 0.9);
  border: 1px solid rgba(239, 68, 68, 0.5);
}

.toast.info {
  background: rgba(59, 130, 246, 0.9);
  border: 1px solid rgba(59, 130, 246, 0.5);
}

/* 动画 */
@keyframes slideIn {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from {
    transform: scale(0.95);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

/* 滚动条美化 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.1);
}

::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .view-container {
    padding: 20px;
  }

  .view-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 20px;
  }

  .header-actions {
    width: 100%;
    justify-content: stretch;
  }

  .header-actions button {
    flex: 1;
  }
  
  .list-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .item-actions {
    width: 100%;
    overflow-x: auto;
    padding-bottom: 4px;
  }
}

.hint-text {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  font-style: italic;
}

.no-changes {
  text-align: center;
  padding: 30px;
  color: #64748b;
  font-size: 14px;
  font-style: italic;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  margin-top: 10px;
}

.change-size {
  flex-shrink: 0;
  font-size: 12px;
  color: #64748b;
  font-family: monospace;
  margin-left: auto;
}
</style>
