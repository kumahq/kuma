import{d as P,o as s,j as c,c as d,w as n,b as u,u as f,e as k,y,D as S,E as x,C as T,G,r as I,f as C,g as M,l,H as E,a as _,z as p,I as L,n as V,K as O,t as r,F as B,J,i as X,A as Q,B as Y}from"./index.8aebf6c5.js";import{_ as R}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.b0a158d7.js";import{E as Z}from"./ErrorBlock.6bd4b5ab.js";import{_ as ee}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.3be41236.js";import{T as ae}from"./TagList.dc62ebc3.js";const te=""+new URL("icon-empty-table.dbb0b754.svg",import.meta.url).href,se={class:"pagination"},ne=P({__name:"PaginationWidget",props:{hasPrevious:{type:Boolean,default:!1},hasNext:{type:Boolean,default:!1}},emits:["next","previous"],setup(t,{emit:g}){const a=t;function D(){g("next"),S.logger.info(x.PAGINATION_NEXT_BUTTON_CLICKED)}function v(){g("previous"),S.logger.info(x.PAGINATION_PREVIOUS_BUTTON_CLICKED)}return(h,N)=>(s(),c("div",se,[a.hasPrevious?(s(),d(f(k),{key:0,appearance:"primary",onClick:v},{default:n(()=>[u(" \u2039 Previous ")]),_:1})):y("",!0),a.hasNext?(s(),d(f(k),{key:1,appearance:"primary",onClick:D},{default:n(()=>[u(" Next \u203A ")]),_:1})):y("",!0)]))}});const oe=T(ne,[["__scopeId","data-v-04fc7cda"]]),w=t=>(Q("data-v-831f8a11"),t=t(),Y(),t),le={class:"data-overview","data-testid":"data-overview"},ie={class:"data-table-controls mb-2"},re=w(()=>l("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 36 36"},[l("g",{fill:"#fff","fill-rule":"nonzero"},[l("path",{d:"M18 5.5a12.465 12.465 0 00-8.118 2.995 1.5 1.5 0 001.847 2.363l.115-.095A9.437 9.437 0 0118 8.5l.272.004a9.487 9.487 0 019.07 7.75l.04.246H25a.5.5 0 00-.416.777l4 6a.5.5 0 00.832 0l4-6 .04-.072A.5.5 0 0033 16.5h-2.601l-.017-.15C29.567 10.2 24.294 5.5 18 5.5zM2.584 18.723l-.04.072A.5.5 0 003 19.5h2.6l.018.15C6.433 25.8 11.706 30.5 18 30.5c3.013 0 5.873-1.076 8.118-2.995a1.5 1.5 0 00-1.847-2.363l-.115.095A9.437 9.437 0 0118 27.5l-.272-.004a9.487 9.487 0 01-9.07-7.75l-.041-.246H11a.5.5 0 00.416-.777l-4-6a.5.5 0 00-.832 0l-4 6z"})])],-1)),ce=[re],de=w(()=>l("span",null,"Refresh",-1)),ue={key:3,class:"data-overview-content"},fe={key:0,class:"data-overview-table"},ge={key:0,class:"action-link__active-state"},ve=w(()=>l("span",{class:"sr-only"},"Selected",-1)),me={key:1,class:"action-link__normal-state"},pe=w(()=>l("div",{class:"card-icon mb-3"},[l("img",{src:te})],-1)),ye={key:0},he={key:1},_e={key:2,class:"data-overview-content mt-6"},be=P({__name:"DataOverview",props:{selectedEntityName:{type:String,required:!1,default:""},pageSize:{type:Number,required:!1,default:12},isLoading:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},isEmpty:{type:Boolean,required:!1,default:!1},emptyState:{type:Object,required:!1,default:null},tableData:{type:Object,required:!1,default:null},tableDataIsEmpty:{type:Boolean,required:!1,default:!1},showWarnings:{type:Boolean,required:!1,default:!1},showDetails:{type:Boolean,required:!1,default:!1},next:{type:[String,Boolean,null],required:!1,default:!1},pageOffset:{type:Number,required:!1,default:0}},emits:["table-action","refresh","load-data"],setup(t,{emit:g}){const a=t,D=G(),v=I(""),h=I(a.pageOffset),N=C(()=>a.showWarnings?a.tableData.headers:a.tableData.headers.filter(o=>o.key!=="warnings")),A=C(()=>a.tableData.headers.map(o=>o.key).filter(o=>D[o])),q=C(()=>`${a.tableData.data.length}-${N.value.length}`);M(()=>a.isLoading,function(){!a.isLoading&&a.tableData.data.length>0&&(v.value=a.selectedEntityName||a.tableData.data[0].name)});function z(){return{data:a.tableData.data,total:a.tableData.data.length}}function $(o,m){v.value=m.name,g("table-action",m)}function K(){g("refresh"),g("load-data",h.value),S.logger.info(x.TABLE_REFRESH_BUTTON_CLICKED)}function U(){h.value=a.pageOffset-a.pageSize,g("load-data",a.pageOffset-a.pageSize)}function W(){h.value=a.pageOffset+a.pageSize,g("load-data",a.pageOffset+a.pageSize)}function H({headerKey:o}){return{class:["warnings"].includes(o)?"text-center":["details"].includes(o)?"text-right":""}}function F({name:o}){const m=a.selectedEntityName||a.tableData.data[0].name;return{class:o===m?"is-selected":""}}return(o,m)=>{const b=X("router-link");return s(),c("div",le,[l("div",ie,[E(o.$slots,"additionalControls",{},void 0,!0),_(f(k),{class:"refresh-button",appearance:"primary",disabled:t.isLoading,onClick:K},{default:n(()=>[l("span",{class:p(["refresh-icon custom-control-icon",{"is-spinning":t.isLoading}])},ce,2),de]),_:1},8,["disabled"])]),t.isLoading?(s(),d(ee,{key:0})):t.error!==null?(s(),d(Z,{key:1,error:t.error},null,8,["error"])):t.isEmpty?(s(),d(R,{key:2})):(s(),c("div",ue,[!t.tableDataIsEmpty&&t.tableData?(s(),c("div",fe,[(s(),d(f(J),{key:f(q),class:p({"data-table-is-hidden":t.tableDataIsEmpty}),fetcher:z,headers:f(N),"cell-attrs":H,"row-attrs":F,"disable-pagination":"","is-clickable":"","data-testid":"data-overview-table","onRow:click":$},L({status:n(({rowValue:e})=>[l("div",{class:p(["entity-status",{"is-offline":e.toLowerCase()==="offline"||e===!1,"is-online":e.toLowerCase()==="online","is-degraded":e.toLowerCase()==="partially degraded","is-not-available":e.toLowerCase()==="not available"}])},[l("span",null,r(e),1)],2)]),tags:n(({rowValue:e})=>[_(ae,{tags:e},null,8,["tags"])]),name:n(({row:e,rowValue:i})=>[e.nameRoute?(s(),d(b,{key:0,to:e.nameRoute},{default:n(()=>[u(r(i),1)]),_:2},1032,["to"])):(s(),c(B,{key:1},[u(r(i),1)],64))]),mesh:n(({row:e,rowValue:i})=>[e.meshRoute?(s(),d(b,{key:0,to:e.meshRoute},{default:n(()=>[u(r(i),1)]),_:2},1032,["to"])):(s(),c(B,{key:1},[u(r(i),1)],64))]),service:n(({row:e,rowValue:i})=>[e.serviceInsightRoute?(s(),d(b,{key:0,to:e.serviceInsightRoute},{default:n(()=>[u(r(i),1)]),_:2},1032,["to"])):(s(),c(B,{key:1},[u(r(i),1)],64))]),totalUpdates:n(({row:e})=>[l("span",null,r(e.totalUpdates),1)]),selected:n(({row:e})=>[l("a",{class:p(["data-table-action-link",{"is-active":v.value===e.name}])},[v.value===e.name?(s(),c("span",ge,[u(" \u2713 "),ve])):(s(),c("span",me," View "))],2)]),dpVersion:n(({row:e,rowValue:i})=>[l("div",{class:p({"with-warnings":e.unsupportedEnvoyVersion||e.unsupportedKumaDPVersion||e.kumaDpAndKumaCpMismatch})},r(i),3)]),envoyVersion:n(({row:e,rowValue:i})=>[l("div",{class:p({"with-warnings":e.unsupportedEnvoyVersion})},r(i),3)]),_:2},[V(f(A),e=>({name:e,fn:n(({rowValue:i,row:j})=>[E(o.$slots,e,{rowValue:i,row:j},void 0,!0)])})),t.showWarnings?{name:"warnings",fn:n(({row:e})=>[e.withWarnings?(s(),d(f(O),{key:0,class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20"})):y("",!0)]),key:"0"}:void 0,t.showDetails?{name:"details",fn:n(({row:e})=>[_(f(k),{class:"detail-link",appearance:"btn-link",to:e.nameRoute},{icon:n(()=>[_(f(O),{icon:e.warnings.length>0?"warning":"info",color:e.warnings.length>0?"var(--black-75)":"var(--blue-500)","secondary-color":e.warnings.length>0?"var(--yellow-300)":void 0,size:"20"},null,8,["icon","color","secondary-color"])]),default:n(()=>[u(" Details ")]),_:2},1032,["to"])]),key:"1"}:void 0]),1032,["class","headers"])),_(oe,{"has-previous":h.value>0,"has-next":Boolean(t.next),onNext:W,onPrevious:U},null,8,["has-previous","has-next"])])):y("",!0),t.tableDataIsEmpty&&t.tableData?(s(),d(R,{key:1},L({title:n(()=>[pe,t.emptyState.title?(s(),c("p",ye,r(t.emptyState.title),1)):(s(),c("p",he," No items found "))]),_:2},[t.emptyState.message?{name:"message",fn:n(()=>[u(r(t.emptyState.message),1)]),key:"0"}:void 0]),1024)):y("",!0),o.$slots.content?(s(),c("div",_e,[E(o.$slots,"content",{},void 0,!0)])):y("",!0)]))])}}});const Ee=T(be,[["__scopeId","data-v-831f8a11"]]);export{Ee as D};
