import Vue from 'vue'
import Vuex from 'vuex'
import { getAllPluginPackageResourceFiles, getMyMenus } from '../api/server'
import { getChildRouters } from '../pages/util/router.js'
import { MENUS } from '../const/menus.js'
Vue.use(Vuex)
const store = new Vuex.Store({
  state: {
    packageFiles: [],
    allMenus: []
  },
  mutations: {
    setMenus (state, value) {
      state.allMenus = value
    },
    setFileData (state, value) {
      state.packageFiles = value
    }
  },
  actions: {
    updateMenus ({ commit }) {
      let menus = []
      getMyMenus().then(({ status, data }) => {
        if (status === 'OK') {
          data.forEach(_ => {
            if (!_.category) {
              let menuObj = MENUS.find(m => m.code === _.code)
              if (menuObj) {
                menus.push({
                  title: Vue.config.lang === 'zh-CN' ? menuObj.cnName : menuObj.enName,
                  id: _.id,
                  submenus: [],
                  ..._,
                  ...menuObj
                })
              } else {
                menus.push({
                  title: _.code,
                  id: _.id,
                  submenus: [],
                  ..._
                })
              }
            }
          })
          data.forEach(_ => {
            if (_.category) {
              let menuObj = MENUS.find(m => m.code === _.code)
              if (menuObj) {
                // Platform Menus
                menus.forEach(h => {
                  if (_.category === '' + h.id) {
                    h.submenus.push({
                      title: Vue.config.lang === 'zh-CN' ? menuObj.cnName : menuObj.enName,
                      id: _.id,
                      ..._,
                      ...menuObj
                    })
                  }
                })
              } else {
                // Plugins Menus
                menus.forEach(h => {
                  if (_.category === '' + h.id) {
                    h.submenus.push({
                      title: Vue.config.lang === 'zh-CN' ? _.localDisplayName : _.displayName,
                      id: _.id,
                      link: _.path,
                      ..._
                    })
                  }
                })
              }
            }
          })
          window.localStorage.setItem('wecube_cache_menus', JSON.stringify(menus))
          commit('setMenus', menus)
          window.myMenus = menus
          getChildRouters(window.routers || [])
        }
      })
    },
    getAllPluginPackageResourceFiles ({ commit }) {
      getAllPluginPackageResourceFiles().then(res => {
        if (res.status === 'OK' && res.data && res.data.length > 0) {
          Vue.prototype.$Notice.info({
            title: Vue.t('notification_desc')
          })
          const eleContain = document.getElementsByTagName('body')
          let script = {}
          commit('setFileData', res.data || [])
          res.data.forEach(file => {
            if (file.relatedPath.indexOf('.js') > -1) {
              let contains = document.createElement('script')
              contains.type = 'text/javascript'
              contains.src = file.relatedPath
              script[file.packageName] = contains
              eleContain[0].appendChild(contains)
            }
            if (file.relatedPath.indexOf('.css') > -1) {
              let contains = document.createElement('link')
              contains.type = 'text/css'
              contains.rel = 'stylesheet'
              contains.href = file.relatedPath
              eleContain[0].appendChild(contains)
            }
          })
          Object.keys(script).forEach(key => {
            if (script[key].readyState) {
              // IE
              script[key].onreadystatechange = () => {
                if (script[key].readyState === 'complete' || script[key].readyState === 'loaded') {
                  script[key].onreadystatechange = null
                }
              }
            } else {
              // Non IE
              script[key].onload = () => {
                setTimeout(() => {
                  Vue.prototype.$Notice.success({
                    title: `${key} ${Vue.t('plugin_load')}`
                  })
                }, 0)
              }
            }
          })
        }
      })
    }
  }
})
export default store
