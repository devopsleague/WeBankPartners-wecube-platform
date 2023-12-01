<template>
  <div id="vue-g6-editor">
    <!-- 工具栏 -->
    <Row>
      <Col :span="24">
        <div id="toolbar">
          <!-- <i data-command="save" class="command fa fa-floppy-o" title="保存"></i> -->

          <!-- <i class="fa fa-history" title="历史数据" @click="readHistoryData"></i>
          <i class="fa fa-hdd-o" title="上传数据" @click="readUploadData"></i>
          <i class="fa fa-download" title="另存为文件" @click="saveAsFile"></i> -->

          <!-- <i class="fa fa-picture-o" title="另存为图片" @click="openSaveAsImageDialog"></i> -->

          <!-- <i data-command="undo" class="command fa fa-undo" title="撤销"></i>
          <i data-command="redo" class="command fa fa-repeat" title="重做"></i> -->
          <!-- <i data-command="delete" class="command fa fa-trash-o" title="删除"></i> -->
          <!-- <i data-command="zoomOut" class="command fa fa-search-minus" title="缩小"></i>
          <i data-command="zoomIn" class="command fa fa-search-plus" title="放大"></i>
          <i data-command="clear" class="command fa fa-eraser" title="清除画布"></i> -->
          <!-- <i data-command="toFront" class="command fa fa-arrow-up" title="提升层级"></i>
          <i data-command="toBack" class="command fa fa-arrow-down" title="下降层级"></i> -->
          <!-- <i data-command="selectAll" class="command fa fa-check-square-o" title="全选"></i> -->
          <!-- <i data-command="copy" class="command fa fa-files-o" title="复制"></i>
          <i data-command="paste" class="command fa fa-clipboard" title="粘贴"></i> -->
          <!-- <i data-command="autoZoom" class="command fa fa-expand" title="实际大小"></i>
          <i data-command="resetZoom" class="command fa fa-compress" title="适应页面"></i> -->
          <!-- <i data-command="addGroup" class="command fa fa-object-group" title="组合"></i>
          <i data-command="unGroup" class="command fa fa-object-ungroup" title="取消组合"></i> -->
          <!-- <i data-command="multiSelect" class="command fa fa fa-crop" title="多选"></i> -->
        </div>
      </Col>
    </Row>
    <!-- 元素面板 + 画布 + 属性栏 -->
    <Row>
      <!-- 元素面板 -->
      <Col :span="2">
        <div id="itempannel">
          <!-- 开始节点 -->
          <div
            id="startNode"
            class="getItem"
            data-type="node"
            data-shape="flow-circle"
            data-size="48*48"
            data-label="开始节点"
            data-color="#FA8C16"
            data-nodeType="startNode"
          >
            <img style="width: 48px; height: 48px" draggable="false" :src="startNodeSVGUrl" alt srcset />
          </div>

          <!-- 常规节点 -->
          <div
            id="regularNode"
            class="getItem"
            data-type="node"
            data-size="100*50"
            data-label="常规节点"
            data-color="#1890ff"
          >
            <img style="width: 70px; height: 40px" draggable="false" :src="regularNodeSVGUrl" alt srcset />
          </div>

          <!-- 条件节点 -->
          <div
            id="judgeNode"
            class="getItem"
            data-type="node"
            data-shape="flow-rhombus"
            data-size="80*80"
            data-label="条件节点"
            data-color="#13C2C2"
          >
            <img style="width: 60px; height: 60px" draggable="false" :src="conditionNodeSVGUrl" />
          </div>

          <!-- 结束节点 -->
          <div
            id="endNode"
            class="getItem"
            data-type="node"
            data-shape="flow-circle"
            data-size="48*48"
            data-label="结束节点"
            data-color="#FA8C16"
            data-nodeType="endNode"
          >
            <img style="width: 48px; height: 48px" draggable="false" :src="endNodeSVGUrl" />
          </div>
        </div>
      </Col>
      <!-- 画布 -->
      <Col :span="22">
        <div id="page"></div>
      </Col>
      <!-- 属性栏 -->
      <Col :span="4">
        <section class="right-part">
          <div id="detailpannel">
            <!-- 节点属性栏 -->
            <div id="nodeAttributeBar" class="pannel" data-status="node-selected">
              <div class="title">节点属性</div>
              <div class="main">
                <Form :model="nodeAttributeForm" label-position="right" :label-width="80">
                  <FormItem label="节点文本">
                    <Input v-model="nodeAttributeForm.label" @on-change="saveNodeAttribute"></Input>
                  </FormItem>
                  <FormItem label="宽度">
                    <InputNumber
                      :max="100"
                      :min="1"
                      v-model="nodeAttributeForm.width"
                      @on-change="saveNodeAttribute"
                    ></InputNumber>
                  </FormItem>
                  <FormItem label="高度">
                    <InputNumber
                      :max="100"
                      :min="1"
                      v-model="nodeAttributeForm.height"
                      @on-change="saveNodeAttribute"
                    ></InputNumber>
                  </FormItem>
                  <FormItem label="颜色">
                    <ColorPicker v-model="nodeAttributeForm.color" @on-change="saveNodeAttribute" />
                  </FormItem>
                </Form>
              </div>
            </div>
            <!-- 边属性栏 -->
            <div id="edgeAttributeBar" class="pannel" data-status="edge-selected">
              <div class="title">边属性</div>
              <div class="main">
                <Form :model="edgeAttributeForm" label-position="right" :label-width="80">
                  <FormItem label="边文本">
                    <Input v-model="edgeAttributeForm.label" @on-change="saveEdgeAttribute"></Input>
                  </FormItem>
                  <FormItem label="边文本">
                    <Select v-model="edgeAttributeForm.shape" @on-change="saveEdgeAttribute">
                      <Option label="流程图折线" value="flow-polyline"></Option>
                      <Option label="流程图圆⻆折线" value="flow-polyline-round"></Option>
                      <Option label="流程图曲线" value="flow-smooth"></Option>
                    </Select>
                  </FormItem>
                </Form>
              </div>
            </div>
            <div id="groupAttributeBar" class="pannel" data-status="group-selected">
              <div class="title">群组属性栏</div>
            </div>
            <div id="canvasAttributeBar" class="pannel" data-status="canvas-selected">
              <div class="title">画布属性栏</div>
              <div class="main">
                <Form :label-width="80" label-position="right">
                  <FormItem label="网格对齐">
                    <Checkbox v-model="canvasAttributeForm.grid" @on-change="toggleGridShowStatus"></Checkbox>
                  </FormItem>
                </Form>
              </div>
            </div>
            <div id="multiAttributeBar" class="pannel" data-status="multi-selected">
              <div class="title">多选时属性栏</div>
            </div>
          </div>
          <!-- 缩略图 -->
          <div id="minimap">
            <div class="title">缩略图</div>
          </div>
        </section>
      </Col>
    </Row>
    <!-- 弹窗 -->
    <article>
      <!-- 下载图片 -->
      <section class="save-as-image-dialog">
        <Modal title="下载图片" v-model="saveAsImageDialogVisible" width="360px">
          <Form :label-width="100" label-position="right">
            <FormItem label="选择图片格式">
              <Select v-model="saveAsImageFormat">
                <Option label="jpg" value="jpg">
                  <span style="float: left">jpg</span>
                  <span style="float: right; color: #8492a6; font-size: 13px">白色背景</span>
                </Option>
                <Option label="png" value="png">
                  <span style="float: left">png</span>
                  <span style="float: right; color: #8492a6; font-size: 13px">透明背景</span>
                </Option>
              </Select>
            </FormItem>
          </Form>
          <span slot="footer">
            <Button @click="saveAsImageDialogVisible = false">取 消</Button>
            <Button type="primary" @click="saveAsImage">确 定</Button>
          </span>
        </Modal>
      </section>
    </article>
    <!-- 右键菜单 -->
    <!-- <section>
      <div id="contextmenu">
        <div data-status="node-selected" class="menu">
          <Button data-command="copy" class="command">复制</Button>
          <Button data-command="paste" class="command">粘贴</Button>
          <Button data-command="delete" class="command">删除</Button>
        </div>
        <div data-status="edge-selected" class="menu">
          <Button data-command="delete" class="command">删除</Button>
        </div>
        <div data-status="group-selected" class="menu">
          <Button data-command="copy" class="command">复制</Button>
          <Button data-command="paste" class="command">粘贴</Button>
          <Button data-command="unGroup" class="command">取消组合</Button>
          <Button data-command="delete" class="command">删除</Button>
        </div>
        <div data-status="canvas-selected" class="menu">
          <Button data-command="undo" class="command">撤销</Button>
          <Button data-command="redo" class="command disable">重做</Button>
        </div>
        <div data-status="multi-selected" class="menu">
          <Button data-command="copy" class="command">复制</Button>
          <Button data-command="paste" class="command">粘贴</Button>
          <Button data-command="addGroup" class="command">组合</Button>
        </div>
      </div>
    </section> -->
  </div>
</template>
<script>
import G6Editor from '@antv/g6-editor'
// import mixin from './mixin'
// import { construct } from 'netflix-conductor-json-tree/dist/index'
export default {
  name: 'VueG6Editor',
  // mixins: [mixin],
  data () {
    return {
      // 节点属性表单
      nodeAttributeForm: {
        label: '',
        width: 0,
        height: 0,
        color: ''
      },
      // 节点属性表单
      edgeAttributeForm: {
        label: ''
      },
      // 画布属性栏表单
      canvasAttributeForm: {
        grid: true,
        cell: 20
      },
      // SVG节点图片URL地址
      startNodeSVGUrl: require('./node-img/start-node.svg'),
      endNodeSVGUrl: require('./node-img/end-node.svg'),
      regularNodeSVGUrl: require('./node-img/regular-node.svg'),
      conditionNodeSVGUrl: require('./node-img/condition-node.svg'),
      modelNodeSVGUrl: 'https://gw.alipayobjects.com/zos/rmsportal/rQMUhHHSqwYsPwjXxcfP.svg',
      // 编辑器
      editor: null,
      saveAsImageDialogVisible: false,
      saveAsImageFormat: 'jpg'
    }
  },
  mounted () {
    this.initG6Editor()
  },
  methods: {
    // 初始化
    initG6Editor () {
      const editor = new G6Editor()
      this.editor = editor
      G6Editor.track(false)
      const Command = G6Editor.Command
      // 注册新命令save
      Command.registerCommand('save', {
        // 禁止保存命令进入队列
        queue: false,
        // 命令是否可用
        enable: editor => {
          return true
        },
        // 正向命令
        execute (editor) {
          let needSaveData = editor.getCurrentPage().save()
          console.log(needSaveData)
          localStorage.setItem('flowData', JSON.stringify(needSaveData))
          this.save(needSaveData)
          this.$Notice.success({
            title: 'Success',
            desc: '数据已保存'
          })
        },
        // 反向命令
        back (editor) {
          console.log('反向命令')
          console.log(editor)
        },
        // 快捷键：Ctrl + S
        shortcutCodes: [
          ['metaKey', 's'],
          ['ctrlKey', 's']
        ]
      })
      // 画布
      const flow = new G6Editor.Flow({
        graph: {
          container: 'page'
        },
        align: {
          line: {
            // 对齐线颜色
            stroke: '#FA8C16',
            // 对齐线粗细
            lineWidth: 1
          },
          // 开启全方位对齐
          item: true,
          // 网格对齐
          grid: true
        },
        grid: {
          // 网孔尺寸
          cell: 18
        },
        shortcut: {
          // 开启自定义命令保存的快捷键
          save: true
        }
      })
      window.flow = flow

      // 设置边
      flow.getGraph().edge({
        shape: 'flow-polyline'
      })

      // 元素面板栏
      const itempannel = new G6Editor.Itempannel({
        container: 'itempannel'
      })
      // 工具栏
      const toolbar = new G6Editor.Toolbar({
        container: 'toolbar'
      })
      // 属性栏
      const detailpannel = new G6Editor.Detailpannel({
        container: 'detailpannel'
      })
      // 缩略图
      let minimapWidth = getComputedStyle(document.querySelector('.right-part')).width
      minimapWidth = Number(minimapWidth.replace(/px$/, ''))
      const minimap = new G6Editor.Minimap({
        container: 'minimap',
        width: minimapWidth,
        height: 200
      })
      // 右键菜单
      // const contextmenu = new G6Editor.Contextmenu({
      //   container: 'contextmenu'
      // })
      // 挂载以上组件到Editor
      editor.add(flow)
      editor.add(itempannel)
      editor.add(toolbar)
      editor.add(detailpannel)
      editor.add(minimap)
      // editor.add(contextmenu)
      // 挂载到window，方便调试
      window.editor = editor

      // 获取当前画布
      const currentPage = editor.getCurrentPage()
      currentPage.on('afterchange', e => {
        if (e.action === 'add') {
          if (e.model.nodetype === 'startNode' || e.model.nodetype === 'endNode') {
            let nodes = this.editor.getCurrentPage().getNodes()
            for (const item of nodes) {
              if (item.model.nodetype === e.model.nodetype && item.model.id !== e.model.id) {
                this.editor.getCurrentPage().remove(e.item)
                this.$message.warning('只能有一个开始节点或结束节点')
              }
            }
          }
        }
      })
      // 监听（选择对象后）事件
      currentPage.on('afteritemselected', ev => {
        console.log('打印所选对象属性', ev.item)
        console.log('打印所选对象数据模型', ev.item.model)
        const selectedItemDataModel = ev.item.model
        // 如果选择的对象是节点
        if (ev.item.isNode) {
          this.nodeAttributeForm.label = selectedItemDataModel.label
          this.nodeAttributeForm.width = Number(selectedItemDataModel.size.split('*')[0])
          this.nodeAttributeForm.height = Number(selectedItemDataModel.size.split('*')[1])
          this.nodeAttributeForm.color = selectedItemDataModel.color
        }
        // 如果选择的对象是边
        if (ev.item.isEdge) {
          ev.item.graph.edge({
            shape: 'flow-polyline-round'
          })
          this.edgeAttributeForm.label = selectedItemDataModel.label
          this.edgeAttributeForm.shape = selectedItemDataModel.shape
        }
      })
      // 监听（删除后）事件
      currentPage.on('afterdelete', ev => {})
    },
    // 打开保存为图片弹窗
    openSaveAsImageDialog () {
      this.saveAsImageDialogVisible = true
    },
    // 开启/关闭网格对齐
    toggleGridShowStatus (value) {
      if (value) {
        this.editor.getCurrentPage().showGrid()
      } else {
        this.editor.getCurrentPage().hideGrid()
      }
    },
    // 保存为图片
    saveAsImage () {
      let newCanvas
      if (this.saveAsImageFormat === 'jpg') {
        let canvas = this.editor.getCurrentPage().saveImage()
        newCanvas = document.createElement('canvas')
        newCanvas.width = canvas.width
        newCanvas.height = canvas.height
        let newContext = newCanvas.getContext('2d')
        newContext.fillStyle = '#fff'
        newContext.fillRect(0, 0, newCanvas.width, newCanvas.height)
        newContext.drawImage(canvas, 0, 0)
      }
      if (this.saveAsImageFormat === 'png') {
        newCanvas = this.editor.getCurrentPage().saveImage()
      }
      let imageDataURL = newCanvas.toDataURL()
      let downloadLink = document.createElement('a')
      downloadLink.download = '图片.jpg'
      downloadLink.href = imageDataURL
      document.body.appendChild(downloadLink)
      downloadLink.click()
      document.body.removeChild(downloadLink)
      this.saveAsImageDialogVisible = false
    },
    // 保存为文件
    saveAsFile () {
      let jsonString = JSON.stringify(this.editor.getCurrentPage().save())
      let blob = new Blob([jsonString])
      let blobURL = URL.createObjectURL(blob)
      let downloadLink = document.createElement('a')
      downloadLink.download = '数据.json'
      downloadLink.href = blobURL
      document.body.appendChild(downloadLink)
      downloadLink.click()
      URL.revokeObjectURL(blobURL)
      document.body.removeChild(downloadLink)
    },
    // 读取历史数据
    readHistoryData () {
      let stringData = localStorage.getItem('flowData')
      if (stringData === '' || stringData === '{}' || stringData === null) {
        this.$message.warning('无历史数据')
        return
      }
      let jsonData = JSON.parse(stringData)
      this.editor.getCurrentPage().read(jsonData)
    },
    // 读取上传数据
    readUploadData () {
      let uploadButton = document.createElement('input')
      uploadButton.setAttribute('type', 'file')
      uploadButton.setAttribute('accept', '.json')
      uploadButton.addEventListener('change', e => {
        console.dir(uploadButton)
        let file = uploadButton.files[0]
        let fileReader = new FileReader()
        fileReader.onload = event => {
          console.log(event)
          let text = JSON.parse(event.target.result)
          console.log(text)
          this.editor.getCurrentPage().read(text)
        }
        fileReader.readAsText(file)
      })
      uploadButton.click()
    },
    // 保存节点属性
    saveNodeAttribute () {
      this.editor.executeCommand(() => {
        // 获取画布
        const page = this.editor.getCurrentPage()
        // 获取所选对象
        const selectedItem = page.getSelected()[0]
        page.update(selectedItem.id, {
          label: this.nodeAttributeForm.label,
          size: this.nodeAttributeForm.width + '*' + this.nodeAttributeForm.height,
          color: this.nodeAttributeForm.color
        })
      })
    },
    // 保存边属性
    saveEdgeAttribute () {
      this.editor.executeCommand(() => {
        // 获取画布
        const page = this.editor.getCurrentPage()
        // 获取所选对象
        const selectedItem = page.getSelected()[0]
        console.log(this.edgeAttributeForm)
        page.update(selectedItem.id, {
          label: this.edgeAttributeForm.label,
          shape: this.edgeAttributeForm.shape
        })
      })
    },
    //
    save (source) {
      console.log(source)
      // let edges = source.edges
      // let nodes = source.nodes
      // console.log(11, construct)
      // let res = construct(source)
      // console.log(22, JSON.stringify(res, null, 2))
    }
  }
}
</script>

<style lang="scss" scoped>
$backgroundColor: #fbfbfb;
$borderColor: #dadce0;
$itempannelAndPageBorder: 1px solid #ccc;
$pageHeight: calc(100vh - 41px - 37px - 70px);

#vue-g6-editor {
  // 工具栏
  #toolbar {
    background: $backgroundColor;
    border-bottom: 1px solid $borderColor;
    padding: 4px 14px;

    i {
      font-size: 18px;
      padding: 4px;
      margin-right: 8px;
      color: #999999;

      &:hover {
        cursor: pointer;
        background-color: #eeeeee;
        color: #5cb6ff;
      }
    }
  }

  // 元素面板
  #itempannel {
    box-sizing: border-box;
    background-color: $backgroundColor;
    border-right: 1px solid $borderColor;
    height: $pageHeight;
    padding-top: 10px;
    overflow: hidden;
    display: flex;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: space-around;
    align-content: flex-start;

    .getItem {
      cursor: move;
      width: 80px;
      height: 80px;
      margin-bottom: 20px;
      display: flex;
      justify-content: center;
      align-items: center;

      img {
        width: 100%;
      }
    }
  }

  // 主画布
  #page {
    height: $pageHeight;

    canvas {
      // 去除canvas底部间隙
      display: block;
      width: 100%;
    }
  }

  // 右侧部分（属性栏 + 缩略图）
  .right-part {
    height: $pageHeight;
    display: flex;
    flex-direction: column;
    justify-content: flex-start;
  }

  // 属性栏
  #detailpannel {
    flex-grow: 1;

    background-color: $backgroundColor;
    border-left: 1px solid $borderColor;
    overflow-y: scroll;

    #nodeAttributeBar,
    #edgeAttributeBar,
    #groupAttributeBar,
    #canvasAttributeBar,
    #multiAttributeBar {
      .title {
        height: 34px;
        line-height: 34px;
        text-align: center;
        box-sizing: border-box;
        font-weight: bold;
        font-size: 13px;
        border-width: 0 0 1px 0;
        border-style: solid;
        border-color: $borderColor;
      }

      .main {
        padding: 10px;
      }
    }
  }

  // 缩略图
  #minimap {
    background-color: $backgroundColor;
    border-top: 1px solid #ccc;
    border-left: 1px solid #ccc;

    .title {
      height: 34px;
      line-height: 34px;
      text-align: center;
      box-sizing: border-box;
      font-weight: bold;
      font-size: 13px;
      border-width: 0 0 1px 0;
      border-style: solid;
      border-color: $borderColor;
    }
  }

  // 右键菜单
  #contextmenu {
    display: none;

    // .menu {
    //   ::v-deep .el-button {
    //     width: 100%;
    //     display: block;
    //     margin-left: 0;
    //     border-radius: 0 !important;
    //     border-bottom: none;

    //     &:nth-last-of-type(1) {
    //       border-bottom: 1px solid #dcdfe6;
    //     }
    //   }
    // }
  }

  // 下载图片弹窗
  .save-as-image-dialog {
    // ::v-deep .el-select {
    //   display: block;
    // }
  }
}
</style>
