<template>
  <div>
    <Header>
      <div class="menus">
        <Menu mode="horizontal" theme="dark">
          <div style="margin-right: 20px">
            <img src="../../assets/logo_WeCube.png" alt="LOGO" @click="goHome" class="img-logo" />
          </div>

          <div v-for="menu in menus" :key="menu.code">
            <MenuItem v-if="menu.submenus.length < 1" :name="menu.title" style="cursor: not-allowed">
              {{ menu.title }}
            </MenuItem>

            <Submenu v-else :name="menu.code">
              <template slot="title">{{ menu.title }}</template>
              <router-link
                v-for="submenu in menu.submenus"
                :key="submenu.code"
                :to="submenu.active ? submenu.link || '' : ''"
              >
                <MenuItem :disabled="!submenu.active" :name="submenu.code">{{ submenu.title }}</MenuItem>
              </router-link>
            </Submenu>
          </div>
        </Menu>
      </div>
      <div class="header-right_container">
        <div class="profile">
          <Dropdown style="cursor: pointer">
            <img class="p-icon" src="../../assets/icon/icon_usr.png" width="12" height="12" />{{ username }}
            <Icon type="ios-arrow-down"></Icon>
            <DropdownMenu slot="list">
              <DropdownItem name="logout" to="/login">
                <a @click="showChangePassword" style="width: 100%; display: block">
                  {{ $t('change_password') }}
                </a>
              </DropdownItem>
              <DropdownItem name="logout" to="/login">
                <a @click="logout" style="width: 100%; display: block">
                  {{ $t('logout') }}
                </a>
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </div>
        <div class="language">
          <Dropdown>
            <a href="javascript:void(0)">
              <img
                class="p-icon"
                v-if="currentLanguage === 'English'"
                src="../../assets/icon/icon_lan_EN.png"
                width="12"
                height="12"
              />
              <img class="p-icon" v-else src="../../assets/icon/icon_lan_CN.png" width="12" height="12" />
              {{ currentLanguage }}
              <Icon type="ios-arrow-down"></Icon>
            </a>
            <DropdownMenu slot="list">
              <DropdownItem v-for="(item, key) in language" :key="item.id" @click.native="changeLanguage(key)">
                {{ item }}
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </div>
        <div class="language">
          <Dropdown>
            <a href="javascript:void(0)">
              <img class="p-icon" src="../../assets/icon/icon_hlp.png" width="12" height="12" />
              {{ $t('help_docs') }}
              <Icon type="ios-arrow-down"></Icon>
            </a>
            <DropdownMenu slot="list">
              <DropdownItem v-for="(item, key) in docs" :key="key" @click.native="changeDocs(item.url)">
                {{ $t(item.name) }}
              </DropdownItem>
            </DropdownMenu>
          </Dropdown>
        </div>
        <div class="version">{{ version }}</div>
      </div>
    </Header>
    <Modal
      v-model="changePassword"
      :title="$t('change_password')"
      :mask-closable="false"
      @on-visible-change="cancelChangePassword"
    >
      <Form ref="formValidate" :model="formValidate" :rules="ruleValidate" :label-width="80">
        <FormItem :label="$t('original_password')" prop="originalPassword">
          <Input
            v-model="formValidate.originalPassword"
            type="password"
            :placeholder="$t('original_password_input_placeholder')"
          ></Input>
        </FormItem>
        <FormItem :label="$t('new_password')" prop="newPassword">
          <Input
            v-model="formValidate.newPassword"
            type="password"
            :placeholder="$t('new_password_input_placeholder')"
          ></Input>
        </FormItem>
        <FormItem :label="$t('confirm_password')" prop="confirmPassword">
          <Input
            v-model="formValidate.confirmPassword"
            type="password"
            :placeholder="$t('confirm_password_input_placeholder')"
          ></Input>
        </FormItem>
      </Form>
      <div slot="footer">
        <Button @click="cancelChangePassword(false)">{{ $t('bc_cancel') }}</Button>
        <Button type="primary" @click="okChangePassword">{{ $t('bc_confirm') }}</Button>
      </div>
    </Modal>
  </div>
</template>
<script>
import Vue from 'vue'
import { getApplicationVersion, changePassword } from '@/api/server.js'
import { clearLocalstorage } from '@/pages/util/localStorage.js'
import { mapState } from 'vuex'
export default {
  data () {
    return {
      username: '',
      currentLanguage: '',
      language: {
        'zh-CN': '简体中文',
        'en-US': 'English'
      },
      docs: [
        {
          name: 'online',
          url: 'wecube_doc_url_online'
        }
        // {
        //   name: 'offline',
        //   url: 'wecube_doc_url_offline'
        // }
      ],
      menus: [],
      needLoad: true,
      version: '',
      changePassword: false,
      formValidate: {
        originalPassword: '',
        newPassword: '',
        confirmPassword: ''
      },
      ruleValidate: {
        originalPassword: [{ required: true, message: 'The Original Password cannot be empty', trigger: 'blur' }],
        newPassword: [{ required: true, message: 'New Password cannot be empty', trigger: 'blur' }],
        confirmPassword: [{ required: true, message: 'Confirm Password cannot be empty', trigger: 'blur' }]
      }
    }
  },
  computed: {
    ...mapState({
      getAllMenus: state => state.allMenus
    })
  },
  watch: {
    getAllMenus: {
      handler (val) {
        if (val) {
          this.menus = val
        }
      },
      immediate: true,
      deep: true
    },
    $lang: async function (lang) {
      await this.$store.dispatch('updateMenus')
      window.location.reload()
    }
  },
  methods: {
    async getApplicationVersion () {
      const { status, data } = await getApplicationVersion()
      if (status === 'OK') {
        this.version = data
        window.localStorage.setItem('wecube_version', this.version)
      } else {
        this.version = window.localStorage.getItem('wecube_version') || ''
      }
    },
    goHome () {
      this.$router.push('/homepage')
    },
    changeDocs (url) {
      window.open(this.$t(url))
    },
    logout () {
      clearLocalstorage()
      window.location.href = window.location.origin + window.location.pathname + '#/login'
    },
    showChangePassword () {
      this.changePassword = true
    },
    okChangePassword () {
      this.$refs['formValidate'].validate(async valid => {
        if (valid) {
          if (this.formValidate.newPassword === this.formValidate.confirmPassword) {
            const { status } = await changePassword(this.formValidate)
            if (status === 'OK') {
              this.$Message.success('Success !')
              this.changePassword = false
            }
          } else {
            this.$Message.warning(this.$t('confirm_password_error'))
          }
        }
      })
    },
    cancelChangePassword (flag = false) {
      if (!flag) {
        this.$refs['formValidate'].resetFields()
        this.changePassword = false
      }
    },
    changeLanguage (lan) {
      Vue.config.lang = lan
      this.currentLanguage = this.language[lan]
      localStorage.setItem('lang', lan)
    },
    getLocalLang () {
      let currentLangKey = localStorage.getItem('lang') || navigator.language
      const lang = this.language[currentLangKey] || 'English'
      this.currentLanguage = lang
    }
  },
  async created () {
    this.getLocalLang()
    this.getApplicationVersion()
    this.username = window.localStorage.getItem('username')
  },
  mounted () {
    // if (window.needReLoad) {
    //   this.getAllPluginPackageResourceFiles()
    //   window.needReLoad = false
    // }
    // this.$eventBusP.$on('updateMenus', () => {
    //   this.getMyMenus()
    // })
  }
}
</script>

<style lang="scss" scoped>
.img-logo {
  height: 20px;
  margin: 0 4px 6px 0;
  vertical-align: middle;
  cursor: pointer;
}
.ivu-layout-header {
  padding: 0 20px;
}
.header {
  display: flex;
  .ivu-layout-header {
    height: 50px;
    line-height: 50px;
    background: linear-gradient(90deg, #8bb8fa 0%, #e1ecfb 100%);
  }
  a {
    color: #404144;
  }
  .menus {
    display: inline-block;
    .ivu-menu-horizontal {
      height: 50px;
      line-height: 50px;
      display: flex;
      .ivu-menu-submenu {
        padding: 0 8px;
        font-size: 15px;
        color: #404144;
      }
      .ivu-menu-item {
        font-size: 15px;
        color: #404144;
      }
    }
    .ivu-menu-dark {
      background: transparent;
    }
    .ivu-menu-dark.ivu-menu-horizontal .ivu-menu-submenu {
      color: #404144;
    }
    .ivu-menu-item-active,
    .ivu-menu-item:hover {
      color: #116ef9;
    }
    .ivu-menu-dark.ivu-menu-horizontal .ivu-menu-submenu-active,
    .ivu-menu-dark.ivu-menu-horizontal .ivu-menu-submenu:hover {
      color: #116ef9;
    }
    .ivu-menu-drop-list {
      .ivu-menu-item-active,
      .ivu-menu-item:hover {
        color: black;
      }
    }
  }
  .header-right_container {
    position: absolute;
    right: 20px;
    top: 0;
    .language,
    .help,
    .version,
    .profile {
      float: right;
      display: inline-block;
      vertical-align: middle;
      margin-left: 20px;
    }
    .version {
      color: #404144;
    }

    .p-icon {
      margin-right: 6px;
    }

    .ivu-dropdown-rel {
      display: flex;
      align-items: center;
      a {
        display: flex;
        align-items: center;
      }
    }
  }
}
</style>
