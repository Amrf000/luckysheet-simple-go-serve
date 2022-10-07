<template>
  <!-- <div style="position: absolute; top: 0">
    <input id="uploadBtn" type="file" @change="loadExcel" /> -->

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

    <!-- <a href="javascript:void(0)" @click="downloadExcel"
      >Download source xlsx file</a
    > -->
  <!-- </div> -->
  <div id="luckysheet"></div>
  <div v-show="isMaskShow" id="tip">Downloading</div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from "vue"; // , defineProps, defineEmits
import { exportExcel } from "./export";
import LuckyExcel from "luckyexcel";
import axios from "axios";
import { Base64 } from "js-base64";
import moment from 'moment'
import { useRoute } from 'vue-router'

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
          console.log("exportJson", window.luckysheet.toJson());
          updateServeData(window.luckysheet.toJson());
        },
      };
      window.luckysheet.create(exportJson);
    })
    .catch(function (error) {
      // 请求失败处理
      console.log(error);
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
    .replace(/[\\]/g, '\\\\')
    .replace(/[\"]/g, '\\\"')
    .replace(/[\/]/g, '\\/')
    .replace(/[\b]/g, '\\b')
    .replace(/[\f]/g, '\\f')
    .replace(/[\n]/g, '\\n')
    .replace(/[\r]/g, '\\r')
    .replace(/[\t]/g, '\\t');
};

const updateServeData = (data) => {
  data = escape(JSON.stringify(data))
  let dat = moment(new Date()).format("YYYY-MM-DDTHH:mm:ssZ");
  axios
    .post("http://127.0.0.1:8080/post", {
      LastMod:  dat,
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
  exportExcel(luckysheet.getAllSheets(), "下载");
};

// !!! create luckysheet after mounted
let socket = null;
// Websoket连接成功事件
const websocketonopen = (res) => {
  console.log("WebSocket连接成功", res);
};
// Websoket接收消息事件
const websocketonmessage = (response) => {
  console.log("数据");
  let tpl = JSON.parse(response.data);
  if(lastMod === tpl.LastMod) {
    console.log("no change")
    return;
  }
  if(tpl.Filename === route.query.filename) {
    console.log("filename not match")
    return
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
  const wsurl = `ws://127.0.0.1:8080/ws?filename=${filename}`;// ?lastMod=${lastMod}
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

const route = useRoute()
onMounted(() => {
  // luckysheet.create({
  //   container: "luckysheet",
  // });
  console.log(route.query)
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
