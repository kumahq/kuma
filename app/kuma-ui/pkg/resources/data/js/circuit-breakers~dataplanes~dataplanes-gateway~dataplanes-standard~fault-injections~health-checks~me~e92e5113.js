(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["circuit-breakers~dataplanes~dataplanes-gateway~dataplanes-standard~fault-injections~health-checks~me~e92e5113"],{"0ada":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",[t.isReady?a("div",{staticClass:"label-list-content"},[t.isLoading||t.isEmpty?t._e():a("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[a("div",{staticClass:"label-list__col-wrapper multi-col"},[t._t("default")],2)]},proxy:!0}],null,!0)})],1):t._e(),t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):t._e(),t.isEmpty&&!t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" There is no data to display. ")]},proxy:!0}],null,!1,1612658095)}):t._e(),t.hasError?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):t._e()],1)},s=[],i={name:"LabelList",props:{items:{type:Object,default:null},title:{type:String,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},computed:{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading}}},r=i,o=(a("d835"),a("2877")),l=Object(o["a"])(r,n,s,!1,null,null,null);e["a"]=l.exports},"1d10":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"component-frame"},[t._t("default")],2)},s=[],i={name:"FrameSkeleton"},r=i,o=(a("a948"),a("2877")),l=Object(o["a"])(r,n,s,!1,null,"666bca0e",null);e["a"]=l.exports},"1d3a":function(t,e,a){"use strict";a.d(e,"a",(function(){return u}));a("b0c0"),a("d3b7"),a("96cf");var n=a("c964"),s=a("f3f3"),i=a("d0ff"),r=a("c6ec");function o(t){return Object(i["a"])(t).sort((function(t,e){return t.name>e.name||t.name===e.name&&t.mesh>e.mesh?1:-1}))}var l=function(t){return 0!==t.total&&t.items&&t.items.length>0?o(t.items):[]};function c(t){var e=t.getSingleEntity,a=t.getAllEntities,n=t.getAllEntitiesFromMesh,i=t.mesh,r=t.query,o=t.size,l=t.offset,c=t.params,u=void 0===c?{}:c,d=Object(s["a"])({size:o,offset:l},u);return e&&r?e({mesh:i,name:r},d):i&&"all"!==i?n&&i?n({mesh:i},d):Promise.resolve():a(d)}function u(t){return d.apply(this,arguments)}function d(){return d=Object(n["a"])(regeneratorRuntime.mark((function t(e){var a,n,s,i,o,u,d,f,p,g,y;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return a=e.getSingleEntity,n=e.getAllEntities,s=e.getAllEntitiesFromMesh,i=e.mesh,o=e.query,u=e.size,d=void 0===u?r["h"]:u,f=e.offset,p=e.params,g=void 0===p?{}:p,t.next=3,c({getSingleEntity:a,getAllEntities:n,getAllEntitiesFromMesh:s,mesh:i,query:o,size:d,offset:f,params:g});case 3:if(y=t.sent,y){t.next=6;break}return t.abrupt("return",{data:[],next:!1});case 6:return t.abrupt("return",{data:y.items?l(y):[y],next:Boolean(y.next)});case 7:case"end":return t.stop()}}),t)}))),d.apply(this,arguments)}},"251b":function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"tab-container",attrs:{"data-testid":"tab-container"}},[t.$slots.tabHeader&&t.isReady?a("header",{staticClass:"tab__header"},[t._t("tabHeader")],2):t._e(),a("div",{staticClass:"tab__content-container"},[t.isReady?a("KTabs",{attrs:{tabs:t.tabs},on:{changed:function(e){return t.switchTab(e)}},scopedSlots:t._u([t._l(t.tabsSlots,(function(e){return{key:e,fn:function(){return[t._t(e)]},proxy:!0}})),{key:"warnings-anchor",fn:function(){return[a("span",{staticClass:"flex items-center with-warnings"},[a("KIcon",{staticClass:"mr-1",attrs:{color:"var(--yellow-400)",icon:"warning","secondary-color":"var(--black-75)",size:"16"}}),a("span",[t._v(" Warnings ")])],1)]},proxy:!0}],null,!0),model:{value:t.tabState,callback:function(e){t.tabState=e},expression:"tabState"}}):t._e(),!0===t.loaders?a("div",[t.isLoading?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),t._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):t._e(),t.hasError?a("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[a("div",{staticClass:"card-icon mb-3"},[a("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):t._e()],1):t._e()],1)])},s=[],i=(a("d81d"),a("ac1f"),a("5319"),a("027b")),r=a("75bb"),o=a("ad12"),l=a.n(o),c={name:"Tabs",components:{KEmptyState:l.a},props:{loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},tabs:{type:Array,required:!0},hasBorder:{type:Boolean,default:!1},initialTabOverride:{type:String,default:null}},data:function(){return{tabState:this.initialTabOverride&&"#".concat(this.initialTabOverride)}},computed:{tabsSlots:function(){return this.tabs.map((function(t){return t.hash.replace("#","")}))},isReady:function(){return!1===this.loaders||!this.isEmpty&&!this.hasError&&!this.isLoading}},methods:{switchTab:function(t){i["a"].logger.info(r["a"].TABS_TAB_CHANGE,{data:{newTab:t}}),this.$emit("onTabChange",t)}}},u=c,d=(a("3be0"),a("f8f8"),a("2877")),f=Object(d["a"])(u,n,s,!1,null,"5f856b5a",null);e["a"]=f.exports},2778:function(t,e,a){"use strict";var n=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("div",{staticClass:"data-overview",attrs:{"data-testid":"data-overview"}},[n("div",{staticClass:"data-table-controls mb-2"},[t._t("additionalControls"),n("KButton",{staticClass:"ml-2 refresh-button",attrs:{appearance:"primary",size:"small",disabled:t.isLoading},on:{click:t.onRefreshButtonClick}},[n("div",{staticClass:"refresh-icon",class:{"is-spinning":t.isLoading}},[n("svg",{attrs:{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 36 36"}},[n("g",{attrs:{fill:"#fff","fill-rule":"nonzero"}},[n("path",{attrs:{d:"M18 5.5a12.465 12.465 0 00-8.118 2.995 1.5 1.5 0 001.847 2.363l.115-.095A9.437 9.437 0 0118 8.5l.272.004a9.487 9.487 0 019.07 7.75l.04.246H25a.5.5 0 00-.416.777l4 6a.5.5 0 00.832 0l4-6 .04-.072A.5.5 0 0033 16.5h-2.601l-.017-.15C29.567 10.2 24.294 5.5 18 5.5zM2.584 18.723l-.04.072A.5.5 0 003 19.5h2.6l.018.15C6.433 25.8 11.706 30.5 18 30.5c3.013 0 5.873-1.076 8.118-2.995a1.5 1.5 0 00-1.847-2.363l-.115.095A9.437 9.437 0 0118 27.5l-.272-.004a9.487 9.487 0 01-9.07-7.75l-.041-.246H11a.5.5 0 00.416-.777l-4-6a.5.5 0 00-.832 0l-4 6z"}})])])]),n("span",[t._v("Refresh")])])],2),t.isReady?n("div",{staticClass:"data-overview-content"},[!t.tableDataIsEmpty&&t.tableData?n("div",{staticClass:"data-overview-table"},[n("KTable",{staticClass:"micro-table",class:{"data-table-is-hidden":t.tableDataIsEmpty,"has-border":t.tableHasBorder},attrs:{options:t.tableDataFiltered,"has-side-border":!1,"has-hover":"","is-clickable":""},on:{"row:click":t.tableRowHandler},scopedSlots:t._u([t._l(t.customSlots,(function(e){return{key:e,fn:function(a){var n=a.rowValue,s=a.row;return[t._t(e,null,{rowValue:n,row:s})]}}})),{key:"status",fn:function(e){var a=e.rowValue;return[n("div",{staticClass:"entity-status",class:{"is-offline":"offline"===a.toString().toLowerCase()||!1===a,"is-degraded":"partially degraded"===a.toString().toLowerCase()||!1===a}},[n("span",{staticClass:"entity-status__dot"}),n("span",{staticClass:"entity-status__label"},[t._v(t._s(a))])])]}},{key:"tags",fn:function(e){var a=e.rowValue;return t._l(a,(function(t,e){return n("EntityTag",{key:e,attrs:{tag:t}})}))}},{key:"totalUpdates",fn:function(e){var a=e.row;return[n("span",{staticClass:"entity-total-updates"},[n("span",[t._v(" "+t._s(a.totalUpdates)+" ")])])]}},{key:"actions",fn:function(e){var a=e.row;return[n("a",{staticClass:"data-table-action-link",class:{"is-active":t.selectedRow===a.name}},[t.selectedRow===a.name?n("span",{staticClass:"action-link__active-state"},[t._v(" ✓ "),n("span",{staticClass:"sr-only"},[t._v(" Selected ")])]):n("span",{staticClass:"action-link__normal-state"},[t._v(" View ")])])]}},{key:"dpVersion",fn:function(e){var a=e.row,s=e.rowValue;return[n("div",{class:{"with-warnings":a.unsupportedEnvoyVersion||a.unsupportedKumaDPVersion||a.kumaDpAndKumaCpMismatch}},[t._v(" "+t._s(s)+" ")])]}},{key:"envoyVersion",fn:function(e){var a=e.row,s=e.rowValue;return[n("div",{class:{"with-warnings":a.unsupportedEnvoyVersion}},[t._v(" "+t._s(s)+" ")])]}},t.showWarnings?{key:"warnings",fn:function(t){var e=t.row;return[e.withWarnings?n("KIcon",{staticClass:"mr-1",attrs:{color:"var(--yellow-400)",icon:"warning","secondary-color":"var(--black-75)",size:"20"}}):n("div")]}}:null],null,!0)}),n("Pagination",{attrs:{"has-previous":t.pageOffset>0,"has-next":t.next},on:{next:t.goToNextPage,previous:t.goToPreviousPage}})],1):t._e(),t.tableDataIsEmpty&&t.tableData?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("img",{attrs:{src:a("a448")}})]),t.emptyState.title?n("span",[t._v(" "+t._s(t.emptyState.title)+" ")]):n("span",[t._v(" No Items Found ")])]},proxy:!0},t.emptyState.message?{key:"message",fn:function(){return[t._v(" "+t._s(t.emptyState.message)+" ")]},proxy:!0}:null],null,!0)}):t._e(),t.$slots.content?n("div",{staticClass:"data-overview-content mt-6"},[t._t("content")],2):t._e()],1):t._e(),t.isLoading?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.25)",size:"42"}})],1),t._v(" Data Loading… ")]},proxy:!0}],null,!1,4031249790)}):t._e(),t.hasError?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:t._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),t._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):t._e()],1)},s=[],i=(a("4de4"),a("d81d"),a("b0c0"),a("a9e3"),a("d0ff")),r=a("027b"),o=a("75bb"),l=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"pagination"},[t.hasPrevious?a("KButton",{ref:"paginatePrev",attrs:{appearance:"primary"},on:{click:t.onPreviousButtonClick}},[t._v(" ‹ Previous ")]):t._e(),t.hasNext?a("KButton",{ref:"paginateNext",attrs:{appearance:"primary"},on:{click:t.onNextButtonClick}},[t._v(" Next › ")]):t._e()],1)},c=[],u={name:"Pagination",props:{hasPrevious:{type:Boolean,default:!1},hasNext:{type:Boolean,default:!1}},methods:{onNextButtonClick:function(){this.$emit("next"),r["a"].logger.info(o["a"].PAGINATION_NEXT_BUTTON_CLICKED)},onPreviousButtonClick:function(){this.$emit("previous"),r["a"].logger.info(o["a"].PAGINATION_PREVIOUS_BUTTON_CLICKED)}}},d=u,f=(a("464e"),a("2877")),p=Object(f["a"])(d,l,c,!1,null,"5c16d41d",null),g=p.exports,y=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("span",{staticClass:"entity-tag"},[a("span",{staticClass:"entity-tag__label",class:"entity-tag__label--"+t.cleanTagLabel(t.tag.label)},[t._v(" "+t._s(t.tag.label)+" ")]),a("span",{staticClass:"entity-tag__value",class:"entity-tag__value--"+t.tag.value},[t._v(" "+t._s(t.tag.value)+" ")])])},b=[],h=(a("ac1f"),a("5319"),{name:"EntityTag",props:{tag:{type:Object,required:!0}},methods:{cleanTagLabel:function(t){return t.toLowerCase().replace(".","-").replace("/","-")}}}),m=h,v=(a("cfa8"),Object(f["a"])(m,y,b,!1,null,"287dac0d",null)),_=v.exports,C={name:"DataOverview",components:{Pagination:g,EntityTag:_},props:{pageSize:{type:Number,default:12},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},emptyState:{type:Object,default:null},tableData:{type:Object,default:null},tableHasBorder:{type:Boolean,required:!1,default:!1},tableDataIsEmpty:{type:Boolean,default:!1},showWarnings:{type:Boolean},next:{type:Boolean,default:!1}},data:function(){return{selectedRow:"",pageOffset:0}},computed:{customSlots:function(){var t=this;return this.tableData.headers.map((function(t){var e=t.key;return e})).filter((function(e){return t.$scopedSlots[e]}))},isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading},tableDataFiltered:function(){var t=this.tableData.data,e=this.tableData.headers,a={headers:e,data:Object(i["a"])(t)};return this.showWarnings||(a.headers=a.headers.filter((function(t){var e=t.key;return"warnings"!==e}))),a}},watch:{isLoading:function(t){!t&&this.tableData.data.length>0&&(this.selectedRow=this.tableData.data[0].name)}},methods:{tableRowHandler:function(t,e,a){this.selectedRow=e.name,this.$emit("tableAction",e)},onRefreshButtonClick:function(){this.$emit("refresh"),this.$emit("loadData",this.pageOffset),r["a"].logger.info(o["a"].TABLE_REFRESH_BUTTON_CLICKED)},goToPreviousPage:function(){this.pageOffset-=this.pageSize,this.$emit("loadData",this.pageOffset)},goToNextPage:function(){this.pageOffset+=this.pageSize,this.$emit("loadData",this.pageOffset)}}},T=C,E=(a("9947"),Object(f["a"])(T,n,s,!1,null,null,null));e["a"]=E.exports},"3be0":function(t,e,a){"use strict";a("7b7d")},"464e":function(t,e,a){"use strict";a("d0b8")},"753a":function(t,e,a){},"75bb":function(t,e,a){"use strict";a.d(e,"a",(function(){return n}));var n={PAGINATION_PREVIOUS_BUTTON_CLICKED:"pagination-previous-button-clicked",PAGINATION_NEXT_BUTTON_CLICKED:"pagination-next-button-clicked",SIDEBAR_ITEM_CLICKED:"sidebar-item-clicked",TABLE_REFRESH_BUTTON_CLICKED:"table-refresh-button-clicked",TABS_TAB_CHANGE:"tabs-tab-change",CREATE_MESH_CLICKED:"create-mesh-clicked",CREATE_DATA_PLANE_PROXY_CLICKED:"create-data-plane-proxy-clicked"}},"7b7d":function(t,e,a){},"82c6":function(t,e,a){},9947:function(t,e,a){"use strict";a("82c6")},a448:function(t,e){t.exports="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI0OCIgaGVpZ2h0PSI0MiI+CiAgPHBhdGggZmlsbD0iI0Q5RDlEOSIgZmlsbC1ydWxlPSJldmVub2RkIiBkPSJNNDggNHY1aC0yVjZIMnYzSDBWMkMwIC44OTU0MzA1Ljg5NTQzMSAwIDIgMGg0NGMxLjEwNDU2OSAwIDIgLjg5NTQzMDUgMiAydjJ6bS0yIDI2aC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDMwdjNIMHYtOGgydjNoNHYySDJ6bTQ0LTEyaC00di0yaDR2LTNoMnY4aC0ydi0zek0yIDE4djNIMHYtOGgydjNoNHYySDJ6bTgtMmg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTEwIDI4aDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptOCAwaDR2MmgtNHYtMnptMTIgMTRoLTR2LTJoNHYtM2gydjNjMCAxLjEwNDU2OTUtLjg5NTQzMSAyLTIgMnpNMiA0MGg0djJIMmMtMS4xMDQ1NjkgMC0yLS44OTU0MzA1LTItMnYtM2gydjN6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6bTggMGg0djJoLTR2LTJ6TTIgMnYyaDQ0VjJIMnoiLz4KPC9zdmc+Cg=="},a710:function(t,e,a){},a948:function(t,e,a){"use strict";a("f9f3")},a9e3:function(t,e,a){"use strict";var n=a("83ab"),s=a("da84"),i=a("94ca"),r=a("6eeb"),o=a("5135"),l=a("c6b6"),c=a("7156"),u=a("c04e"),d=a("d039"),f=a("7c73"),p=a("241c").f,g=a("06cf").f,y=a("9bf2").f,b=a("58a8").trim,h="Number",m=s[h],v=m.prototype,_=l(f(v))==h,C=function(t){var e,a,n,s,i,r,o,l,c=u(t,!1);if("string"==typeof c&&c.length>2)if(c=b(c),e=c.charCodeAt(0),43===e||45===e){if(a=c.charCodeAt(2),88===a||120===a)return NaN}else if(48===e){switch(c.charCodeAt(1)){case 66:case 98:n=2,s=49;break;case 79:case 111:n=8,s=55;break;default:return+c}for(i=c.slice(2),r=i.length,o=0;o<r;o++)if(l=i.charCodeAt(o),l<48||l>s)return NaN;return parseInt(i,n)}return+c};if(i(h,!m(" 0o1")||!m("0b1")||m("+0x1"))){for(var T,E=function(t){var e=arguments.length<1?0:t,a=this;return a instanceof E&&(_?d((function(){v.valueOf.call(a)})):l(a)!=h)?c(new m(C(e)),a,E):C(e)},w=n?p(m):"MAX_VALUE,MIN_VALUE,NaN,NEGATIVE_INFINITY,POSITIVE_INFINITY,EPSILON,isFinite,isInteger,isNaN,isSafeInteger,MAX_SAFE_INTEGER,MIN_SAFE_INTEGER,parseFloat,parseInt,isInteger,fromString,range".split(","),I=0;w.length>I;I++)o(m,T=w[I])&&!o(E,T)&&y(E,T,g(m,T));E.prototype=v,v.constructor=E,r(s,h,E)}},c89c:function(t,e,a){},cfa8:function(t,e,a){"use strict";a("c89c")},d0b8:function(t,e,a){},d835:function(t,e,a){"use strict";a("a710")},f8f8:function(t,e,a){"use strict";a("753a")},f9f3:function(t,e,a){}}]);