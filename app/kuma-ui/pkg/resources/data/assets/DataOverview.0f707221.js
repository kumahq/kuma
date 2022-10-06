import{_ as P,u as R,x as B,y as N,r as _,o as a,e as c,c as d,w as n,b as u,k as m,z as K,A,f as o,q as k,a as v,n as h,B as C,h as z,t as r,F as w,p as W,l as H}from"./index.59498396.js";import{_ as U}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.ca9e3849.js";import{E as q}from"./ErrorBlock.9fdf1c69.js";import{_ as F}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.49103b8d.js";const V={name:"PaginationWidget",components:{KButton:R},props:{hasPrevious:{type:Boolean,default:!1},hasNext:{type:Boolean,default:!1}},emits:["next","previous"],methods:{onNextButtonClick(){this.$emit("next"),B.logger.info(N.PAGINATION_NEXT_BUTTON_CLICKED)},onPreviousButtonClick(){this.$emit("previous"),B.logger.info(N.PAGINATION_PREVIOUS_BUTTON_CLICKED)}}},M={class:"pagination"};function j(t,f,s,x,y,l){const g=_("KButton");return a(),c("div",M,[s.hasPrevious?(a(),d(g,{key:0,ref:"paginatePrev",appearance:"primary",onClick:l.onPreviousButtonClick},{default:n(()=>[u(" \u2039 Previous ")]),_:1},8,["onClick"])):m("",!0),s.hasNext?(a(),d(g,{key:1,ref:"paginateNext",appearance:"primary",onClick:l.onNextButtonClick},{default:n(()=>[u(" Next \u203A ")]),_:1},8,["onClick"])):m("",!0)])}const G=P(V,[["render",j],["__scopeId","data-v-bb9c78f2"]]),X=""+new URL("icon-empty-table.dbb0b754.svg",import.meta.url).href;const J={name:"DataOverview",components:{PaginationWidget:G,EmptyBlock:U,ErrorBlock:q,LoadingBlock:F,KButton:R,KIcon:K,KTable:A},props:{selectedEntityName:{type:String,required:!1,default:""},pageSize:{type:Number,default:12},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},error:{type:[Error,null],required:!1,default:null},isEmpty:{type:Boolean,default:!1},emptyState:{type:Object,default:null},tableData:{type:Object,default:null},tableDataIsEmpty:{type:Boolean,default:!1},showWarnings:{type:Boolean,required:!1,default:!1},showDetails:{type:Boolean,required:!1,default:!1},next:{type:Boolean,default:!1},pageOffset:{type:Number,required:!1,default:0}},emits:["table-action","refresh","load-data"],data(){return{selectedRow:"",internalPageOffset:0}},computed:{customSlots(){return this.tableData.headers.map(({key:t})=>t).filter(t=>this.$slots[t])},tableHeaders(){return this.showWarnings?this.tableData.headers:this.tableData.headers.filter(({key:t})=>t!=="warnings")},tableRecompuationKey(){return`${this.tableData.data.length}-${this.tableHeaders.length}`}},watch:{isLoading(t){!t&&this.tableData.data.length>0&&(this.selectedRow=this.selectedEntityName||this.tableData.data[0].name)}},created(){this.internalPageOffset=this.pageOffset},methods:{tableDataFetcher(){return{data:this.tableData.data,total:this.tableData.data.length}},tableRowHandler(t,f){this.selectedRow=f.name,this.$emit("table-action",f)},onRefreshButtonClick(){this.$emit("refresh"),this.$emit("load-data",this.internalPageOffset),B.logger.info(N.TABLE_REFRESH_BUTTON_CLICKED)},goToPreviousPage(){this.internalPageOffset=this.pageOffset-this.pageSize,this.$emit("load-data",this.pageOffset-this.pageSize)},goToNextPage(){this.internalPageOffset=this.pageOffset+this.pageSize,this.$emit("load-data",this.pageOffset+this.pageSize)},getCellAttributes({headerKey:t}){return{class:["warnings"].includes(t)?"text-center":["details"].includes(t)?"text-right":""}},getRowAttributes({name:t}){const f=this.selectedEntityName||this.tableData.data[0].name;return{class:t===f?"is-selected":""}}}},p=t=>(W("data-v-76f03f1a"),t=t(),H(),t),Q={class:"data-overview","data-testid":"data-overview"},Y={class:"data-table-controls mb-2"},Z=p(()=>o("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 36 36"},[o("g",{fill:"#fff","fill-rule":"nonzero"},[o("path",{d:"M18 5.5a12.465 12.465 0 00-8.118 2.995 1.5 1.5 0 001.847 2.363l.115-.095A9.437 9.437 0 0118 8.5l.272.004a9.487 9.487 0 019.07 7.75l.04.246H25a.5.5 0 00-.416.777l4 6a.5.5 0 00.832 0l4-6 .04-.072A.5.5 0 0033 16.5h-2.601l-.017-.15C29.567 10.2 24.294 5.5 18 5.5zM2.584 18.723l-.04.072A.5.5 0 003 19.5h2.6l.018.15C6.433 25.8 11.706 30.5 18 30.5c3.013 0 5.873-1.076 8.118-2.995a1.5 1.5 0 00-1.847-2.363l-.115.095A9.437 9.437 0 0118 27.5l-.272-.004a9.487 9.487 0 01-9.07-7.75l-.041-.246H11a.5.5 0 00.416-.777l-4-6a.5.5 0 00-.832 0l-4 6z"})])],-1)),$=[Z],ee=p(()=>o("span",null,"Refresh",-1)),te={key:3,class:"data-overview-content"},ae={key:0,class:"data-overview-table"},se=p(()=>o("span",{class:"entity-status__dot"},null,-1)),ne={class:"entity-status__label"},oe={key:0,class:"action-link__active-state"},ie=p(()=>o("span",{class:"sr-only"},"Selected",-1)),le={key:1,class:"action-link__normal-state"},re=p(()=>o("div",{class:"card-icon mb-3"},[o("img",{src:X})],-1)),ce={key:0},de={key:1},ue={key:2,class:"data-overview-content mt-6"};function _e(t,f,s,x,y,l){const g=_("KButton"),O=_("LoadingBlock"),S=_("ErrorBlock"),D=_("EmptyBlock"),b=_("router-link"),E=_("KIcon"),I=_("KTable"),L=_("PaginationWidget");return a(),c("div",Q,[o("div",Y,[k(t.$slots,"additionalControls",{},void 0,!0),v(g,{class:"refresh-button",appearance:"primary",disabled:s.isLoading,onClick:l.onRefreshButtonClick},{default:n(()=>[o("span",{class:h(["refresh-icon custom-control-icon",{"is-spinning":s.isLoading}])},$,2),ee]),_:1},8,["disabled","onClick"])]),s.isLoading?(a(),d(O,{key:0})):s.hasError?(a(),d(S,{key:1,error:s.error},null,8,["error"])):s.isEmpty?(a(),d(D,{key:2})):(a(),c("div",te,[!s.tableDataIsEmpty&&s.tableData?(a(),c("div",ae,[(a(),d(I,{key:l.tableRecompuationKey,class:h({"data-table-is-hidden":s.tableDataIsEmpty}),fetcher:l.tableDataFetcher,headers:l.tableHeaders,"cell-attrs":l.getCellAttributes,"row-attrs":l.getRowAttributes,"disable-pagination":"","is-clickable":"","data-testid":"data-overview-table","onRow:click":l.tableRowHandler},C({status:n(({rowValue:e})=>[o("div",{class:h(["entity-status",{"is-offline":e.toString().toLowerCase()==="offline"||e===!1,"is-degraded":e.toString().toLowerCase()==="partially degraded"||e===!1}])},[se,o("span",ne,r(e),1)],2)]),name:n(({row:e,rowValue:i})=>[e.nameRoute?(a(),d(b,{key:0,to:e.nameRoute},{default:n(()=>[u(r(i),1)]),_:2},1032,["to"])):(a(),c(w,{key:1},[u(r(i),1)],64))]),mesh:n(({row:e,rowValue:i})=>[e.meshRoute?(a(),d(b,{key:0,to:e.meshRoute},{default:n(()=>[u(r(i),1)]),_:2},1032,["to"])):(a(),c(w,{key:1},[u(r(i),1)],64))]),service:n(({row:e,rowValue:i})=>[e.serviceInsightRoute?(a(),d(b,{key:0,to:e.serviceInsightRoute},{default:n(()=>[u(r(i),1)]),_:2},1032,["to"])):(a(),c(w,{key:1},[u(r(i),1)],64))]),totalUpdates:n(({row:e})=>[o("span",null,r(e.totalUpdates),1)]),selected:n(({row:e})=>[o("a",{class:h(["data-table-action-link",{"is-active":y.selectedRow===e.name}])},[y.selectedRow===e.name?(a(),c("span",oe,[u(" \u2713 "),ie])):(a(),c("span",le," View "))],2)]),dpVersion:n(({row:e,rowValue:i})=>[o("div",{class:h({"with-warnings":e.unsupportedEnvoyVersion||e.unsupportedKumaDPVersion||e.kumaDpAndKumaCpMismatch})},r(i),3)]),envoyVersion:n(({row:e,rowValue:i})=>[o("div",{class:h({"with-warnings":e.unsupportedEnvoyVersion})},r(i),3)]),_:2},[z(l.customSlots,e=>({name:e,fn:n(({rowValue:i,row:T})=>[k(t.$slots,e,{rowValue:i,row:T},void 0,!0)])})),s.showWarnings?{name:"warnings",fn:n(({row:e})=>[e.withWarnings?(a(),d(E,{key:0,class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20"})):m("",!0)]),key:"0"}:void 0,s.showDetails?{name:"details",fn:n(({row:e})=>[v(g,{class:"detail-link",appearance:"btn-link",to:e.nameRoute},{icon:n(()=>[v(E,{icon:e.warnings.length>0?"warning":"info",color:e.warnings.length>0?"var(--black-75)":"var(--blue-500)","secondary-color":e.warnings.length>0?"var(--yellow-300)":null,size:"20"},null,8,["icon","color","secondary-color"])]),default:n(()=>[u(" Details ")]),_:2},1032,["to"])]),key:"1"}:void 0]),1032,["class","fetcher","headers","cell-attrs","row-attrs","onRow:click"])),v(L,{"has-previous":y.internalPageOffset>0,"has-next":s.next,onNext:l.goToNextPage,onPrevious:l.goToPreviousPage},null,8,["has-previous","has-next","onNext","onPrevious"])])):m("",!0),s.tableDataIsEmpty&&s.tableData?(a(),d(D,{key:1},C({title:n(()=>[re,s.emptyState.title?(a(),c("p",ce,r(s.emptyState.title),1)):(a(),c("p",de," No items found "))]),_:2},[s.emptyState.message?{name:"message",fn:n(()=>[u(r(s.emptyState.message),1)]),key:"0"}:void 0]),1024)):m("",!0),t.$slots.content?(a(),c("div",ue,[k(t.$slots,"content",{},void 0,!0)])):m("",!0)]))])}const pe=P(J,[["render",_e],["__scopeId","data-v-76f03f1a"]]);export{pe as D};
