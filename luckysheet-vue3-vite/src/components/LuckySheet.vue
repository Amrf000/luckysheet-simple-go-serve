<template>
  <a-modal
    v-model:visible="visible"
    title="共享文件名"
    cancel-text="取消"
    ok-text="确定"
    @ok="handleSubmit"
  >
    <a-form ref="formRef" :model="formData" autocomplete="off">
      <a-form-item
        :label-col="labelCol"
        :wrapper-col="wrapperCol"
        name="filename"
        label="共享文件名"
        :rules="[{ required: true, min: 3, message: '请至少输入3个字符' }]"
      >
        <a-input
          v-model:value="formData.filename"
          placeholder="输入共享文件名"
        />
      </a-form-item>
    </a-form>
  </a-modal>

  <div
    style="
      display: blocked;
      position: absolute;
      top: 5px;
      left: 300px;
      z-index: 100;
      border: 1px solid #333;
      padding: 5px;
    "
  >
    <button style="display: inline-block; margin: 6px" @click="newShare">
      新的共享
    </button>

    <button
      style="display: inline-block; margin: 6px"
      onclick="document.getElementById('uploadBtn').click()"
    >
      加载本地文件
    </button>
    <input
      id="uploadBtn"
      type="file"
      @change="loadExcel"
      style="display: none"
    />

    <!-- <span>Or Load remote xlsx file:</span>

    <select v-model="selected" @change="selectExcel">
      <option disabled value="">Choose</option>
      <option
        v-for="option in options"
        :key="option.text"
        :value="option.value"
      >
        {{ option.text }}
      </option>
    </select> -->

    <button style="display: inline-block; margin: 6px" @click="downloadExcel">
      保存到本地
    </button>
    <button style="display: inline-block; margin: 6px" @click="onlineSave">
      在线保存
    </button>
  </div>
  <div id="luckysheet"></div>
  <div v-show="isMaskShow" id="tip">Downloading</div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue"; // , defineProps, defineEmits
import { exportExcel } from "./export";
import LuckyExcel from "luckyexcel";
import axios from "axios";
import { Base64 } from "js-base64";
import moment from "moment";
import { useRoute, useRouter} from "vue-router";
import { message } from "ant-design-vue";

const isMaskShow = ref(false);
const selected = ref("");
const jsonData = ref({});
const options = ref([
  {
    text: "Money Manager.xlsx",
    value: "https://minio.cnbabylon.com/public/luckysheet/money-manager-2.xlsx",
  },
  {
    text: "Activity costs tracker.xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/Activity%20costs%20tracker.xlsx",
  },
  {
    text: "House cleaning checklist.xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/House%20cleaning%20checklist.xlsx",
  },
  {
    text: "Student assignment planner.xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/Student%20assignment%20planner.xlsx",
  },
  {
    text: "Credit card tracker.xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/Credit%20card%20tracker.xlsx",
  },
  {
    text: "Blue timesheet.xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/Blue%20timesheet.xlsx",
  },
  {
    text: "Student calendar (Mon).xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/Student%20calendar%20%28Mon%29.xlsx",
  },
  {
    text: "Blue mileage and expense report.xlsx",
    value:
      "https://minio.cnbabylon.com/public/luckysheet/Blue%20mileage%20and%20expense%20report.xlsx",
  },
]);

const visible = ref(false);
const newShare = () => {
  visible.value = !visible.value;
};
const rules = ref({
  filename: [
    "filename",
    {
      rules: [{ required: true, min: 3, message: "请至少输入3个字符" }],
      getValueFromEvent: (e) => e.target.value.trim(),
    },
  ],
});
const labelCol = ref({
  xs: { span: 24 },
  sm: { span: 5 },
});
const wrapperCol = ref({
  xs: { span: 24 },
  sm: { span: 16 },
});
const formData = ref({
  filename: "",
});
const formRef = ref();
const handleSubmit = () => {
  formRef.value
    .validateFields()
    .then((values) => {
      let dat = moment(new Date()).format("YYYY-MM-DDTHH:mm:ssZ");
      let filename = formData.value.filename;
      axios
        .post("http://127.0.0.1:8080/create", {
          LastMod: dat,
          Filename: filename,
          Options: "",
        })
        .then((res) => {
          lastMod = dat;
          window.location.href = `${window.location.origin}?filename=${filename}`;
        })
        .catch((error) => {
          message.error("创建失败,共享文件已存在");
        });
      visible.value = false;
      formRef.value.resetFields();
    })
    .catch((info) => {});

  // const response = await fetch("https://some.api/process-form", {
  //   method: "POST",
  //   body: _formData,
  // });
};

const loadExcel = (evt) => {
  const files = evt.target.files;
  if (files == null || files.length == 0) {
    alert("No files wait for import");
    return;
  }

  let name = files[0].name;
  let suffixArr = name.split("."),
    suffix = suffixArr[suffixArr.length - 1];
  if (suffix != "xlsx") {
    alert("Currently only supports the import of xlsx files");
    return;
  }
  LuckyExcel.transformExcelToLucky(
    files[0],
    function (exportJson, luckysheetfile) {
      if (exportJson.sheets == null || exportJson.sheets.length == 0) {
        alert(
          "Failed to read the content of the excel file, currently does not support xls files!"
        );
        return;
      }
      console.log("exportJson", exportJson);
      jsonData.value = exportJson;

      window?.luckysheet?.destroy && window.luckysheet.destroy();
      window.luckysheet.create({
        container: "luckysheet", //luckysheet is the container id
        showinfobar: false,
        data: exportJson.sheets,
        title: exportJson.info.name,
        userInfo: exportJson.info.name.creator,
      });
    }
  );
};

const onlineSave = () => {
  updateServeData(window.luckysheet.toJson());
};

let lastMod = 0;
const getInitData = (filename) => {
  axios
    .get(`http://127.0.0.1:8080/?filename=${filename}`)
    .then((response) => {
      let tpl =
        response.data instanceof String
          ? JSON.parse(response.data)
          : response.data;
      lastMod = tpl.LastMod;
      const exportJson = JSON.parse(
        eval("`" + Base64.decode(tpl.Options) + "`")
      );
      jsonData.value = exportJson;
      window?.luckysheet?.destroy && window.luckysheet.destroy();
      /*
{
        container: "luckysheet", //luckysheet is the container id
        showinfobar: false,
        data: exportJson.sheets,
        title: exportJson.info.name,
        userInfo: exportJson.info.name.creator,
        hook: {
            updated: function (operate) {
               console.log("exportJson",  window.luckysheet.toJson())
               updateServeData(JSON.stringify(window.luckysheet.toJson()))
            }
        }
      }
*/
      exportJson["hook"] = {
        updated: function (operate) {
          updateServeData(window.luckysheet.toJson());
        },
      };
      window.luckysheet.create(exportJson);
    })
    .catch(function (error) {
      // 请求失败处理
      message.error("共享文件不存在");
    });
  /*window.luckysheet.create({
    container: "luckysheet",
    lang: "zh",
    allowUpdate: true,
    updateImageUrl: "http://127.0.0.1:8080/luckysheet/api/updateImg",
    updateUrl: "ws://127.0.0.1:8080/ws",//luckysheet/websocket/luckysheet
    gridKey: "0",
    loadUrl: "http://127.0.0.1:8080",///luckysheet/api/load
    loadSheetUrl: "http://127.0.0.1:8080/luckysheet/api/loadsheet",
  });*/
};

const escape = function (str) {
  return str
    .replace(/[\\]/g, "\\\\")
    .replace(/[\"]/g, '\\"')
    .replace(/[\/]/g, "\\/")
    .replace(/[\b]/g, "\\b")
    .replace(/[\f]/g, "\\f")
    .replace(/[\n]/g, "\\n")
    .replace(/[\r]/g, "\\r")
    .replace(/[\t]/g, "\\t");
};

const updateServeData = (data) => {
  data = escape(JSON.stringify(data));
  let dat = moment(new Date()).format("YYYY-MM-DDTHH:mm:ssZ");
  axios
    .post("http://127.0.0.1:8080/update", {
      LastMod: dat,
      Filename: route.query.filename,
      Options: Base64.encode(data),
    })
    .then((res) => {
      lastMod = dat;
    })
    .catch((error) => {
      console.log(error);
    });
};

const updateClientData = (exportJson) => {
  exportJson["hook"] = {
    updated: function (operate) {
      console.log("exportJson", window.luckysheet.toJson());
      updateServeData(window.luckysheet.toJson());
    },
  };

  jsonData.value = exportJson;

  window?.luckysheet?.destroy && window.luckysheet.destroy();
  window.luckysheet.create(exportJson);
};

const selectExcel = (evt) => {
  const value = selected.value;
  const name = evt.target.options[evt.target.selectedIndex].innerText;

  if (value == "") {
    return;
  }
  isMaskShow.value = true;

  LuckyExcel.transformExcelToLuckyByUrl(
    value,
    name,
    (exportJson, luckysheetfile) => {
      if (exportJson.sheets == null || exportJson.sheets.length == 0) {
        alert(
          "Failed to read the content of the excel file, currently does not support xls files!"
        );
        return;
      }
      console.log("exportJson", exportJson);
      jsonData.value = exportJson;

      isMaskShow.value = false;

      window?.luckysheet?.destroy && window.luckysheet.destroy();
      window.luckysheet.create({
        container: "luckysheet", //luckysheet is the container id
        showinfobar: false,
        data: exportJson.sheets,
        title: exportJson.info.name,
        userInfo: exportJson.info.name.creator,
      });
    }
  );
};
const downloadExcel = () => {
  // const value = selected.value;;
  //
  // if(value.length==0){
  //     alert("Please select a demo file");
  //     return;
  // }
  //
  // var elemIF = document.getElementById("Lucky-download-frame");
  // if(elemIF==null){
  //     elemIF = document.createElement("iframe");
  //     elemIF.style.display = "none";
  //     elemIF.id = "Lucky-download-frame";
  //     document.body.appendChild(elemIF);
  // }
  // elemIF.src = value;
  exportExcel(luckysheet.getAllSheets(), jsonData.value.title);
};

// !!! create luckysheet after mounted
let socket = null;
// Websoket连接成功事件
const websocketonopen = (res) => {
  console.log("WebSocket连接成功", res);
};
// Websoket接收消息事件
const websocketonmessage = (response) => {
  // console.log("数据");
  let tpl = JSON.parse(response.data);
  if (tpl.Filename !== route.query.filename) {
    // console.log("filename not match")
    return;
  }

  if (lastMod === tpl.LastMod) {
    // console.log("no change")
    return;
  }

  const exportJson = JSON.parse(eval("`" + Base64.decode(tpl.Options) + "`"));
  lastMod = tpl.LastMod;
  updateClientData(exportJson);
};
// Websoket连接错误事件
const websocketonerror = (res) => {
  console.log("连接错误", res);
};
// Websoket断开事件
const websocketclose = (res) => {
  console.log("断开连接", res);
};
const setupWs = (filename) => {
  const wsurl = `ws://127.0.0.1:8080/ws?filename=${filename}`; // ?lastMod=${lastMod}
  socket = new WebSocket(wsurl);
  socket.onopen = websocketonopen;
  socket.onmessage = websocketonmessage;
  socket.onerror = websocketonerror;
  socket.onclose = websocketclose;
};

// 组件被销毁之前，清空 sock 对象
onBeforeUnmount(() => {
  // 关闭连接
  websocketclose();

  // 销毁 websocket 实例对象
  socket = null;
});

const route = useRoute();
const router = useRouter()
onMounted(() => {
  // luckysheet.create({
  //   container: "luckysheet",
  // });
  if (!route.query.filename) {
    router.push({ path: route.path, query: { filename: "demo.json" } });
  }
  getInitData(route.query.filename);
  setupWs(route.query.filename);
});

// const props = defineProps({
//   data: {
//     type: Object,
//     default: () => ({}),
//   },
// });

// const emit = defineEmits(['close', 'submit']);
// const onCancel = () => {
//   emit('close');
// };
</script>

<style scoped>
#luckysheet {
  margin: 0px;
  padding: 0px;
  position: absolute;
  width: 100%;
  left: 0px;
  top: 0px;
  bottom: 0px;
}

#uploadBtn {
  font-size: 16px;
}

#tip {
  position: absolute;
  z-index: 1000000;
  left: 0px;
  top: 0px;
  bottom: 0px;
  right: 0px;
  background: rgba(255, 255, 255, 0.8);
  text-align: center;
  font-size: 40px;
  align-items: center;
  justify-content: center;
  display: flex;
}
</style>
