import{d as I,o as s,c as u,i as c,w as n,a as g,u as f,m as k,b as o,O as w,j as p,ce as R,cf as C,k as L,cg as F,bP as P,g as S,h as H,cc as G,e as b,c0 as E,ch as O,cd as Q,F as y,bV as d,ci as B,cj as X,bX as M,bY as Y}from"./index.bd548025.js";import{_ as x}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.0d00632b.js";import{E as J}from"./ErrorBlock.ee0cc1df.js";import{_ as Z}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.f0102383.js";import{S as ee}from"./StatusBadge.fafdc81c.js";import{T as te}from"./TagList.4378af73.js";const we={get(a){const e=new URL(window.location.href).searchParams.get(a);return e!==null?e.replaceAll("+"," "):null},set(a,r){const e=new URL(window.location.href);r!=null?e.searchParams.set(a,String(r).replace(/\s/g,"+")):e.searchParams.has(a)&&e.searchParams.delete(a),window.history.replaceState({path:e.href},"",e.href)}},ae=""+new URL("icon-empty-table.dbb0b754.svg",import.meta.url).href,se={class:"pagination"},ne=I({__name:"PaginationWidget",props:{hasPrevious:{type:Boolean,default:!1},hasNext:{type:Boolean,default:!1}},emits:["next","previous"],setup(a,{emit:r}){const e=a;function D(){r("next"),R.logger.info(C.PAGINATION_NEXT_BUTTON_CLICKED)}function _(){r("previous"),R.logger.info(C.PAGINATION_PREVIOUS_BUTTON_CLICKED)}return(v,N)=>(s(),u("div",se,[e.hasPrevious?(s(),c(f(w),{key:0,appearance:"primary","data-testid":"pagination-previous-button",onClick:_},{default:n(()=>[g(f(k),{icon:"chevronLeft",color:"currentColor",size:"16","hide-title":"","aria-hidden":"true"}),o(`

      Previous
    `)]),_:1})):p("",!0),o(),e.hasNext?(s(),c(f(w),{key:1,appearance:"primary","data-testid":"pagination-next-button",onClick:D},{default:n(()=>[o(`
      Next

      `),g(f(k),{icon:"chevronRight",color:"currentColor",size:"16","hide-title":"","aria-hidden":"true"})]),_:1})):p("",!0)]))}});const oe=L(ne,[["__scopeId","data-v-94d7b089"]]),ie=a=>(M("data-v-ae85a567"),a=a(),Y(),a),le={class:"data-overview","data-testid":"data-overview"},re={class:"data-table-controls mb-2"},de={key:3,class:"data-overview-content"},ce={key:0,class:"data-overview-table"},ue=ie(()=>b("img",{class:"mb-3",src:ae},null,-1)),fe={key:0},ge={key:1},me={key:2,class:"data-overview-content mt-6"},pe=I({__name:"DataOverview",props:{selectedEntityName:{type:String,required:!1,default:""},pageSize:{type:Number,required:!1,default:12},isLoading:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},isEmpty:{type:Boolean,required:!1,default:!1},emptyState:{type:Object,required:!1,default:null},tableData:{type:Object,required:!1,default:null},tableDataIsEmpty:{type:Boolean,required:!1,default:!1},showWarnings:{type:Boolean,required:!1,default:!1},showDetails:{type:Boolean,required:!1,default:!1},next:{type:[String,Boolean,null],required:!1,default:!1},pageOffset:{type:Number,required:!1,default:0}},emits:["table-action","refresh","load-data"],setup(a,{emit:r}){const e=a,D=F(),_=P(""),v=P(e.pageOffset),N=S(()=>e.showWarnings?e.tableData.headers:e.tableData.headers.filter(l=>l.key!=="warnings")),T=S(()=>e.tableData.headers.map(l=>l.key).filter(l=>D[l])),q=S(()=>`${e.tableData.data.length}-${N.value.length}`);H(()=>e.isLoading,function(){!e.isLoading&&e.tableData.data.length>0&&(_.value=e.selectedEntityName||e.tableData.data[0].name)});function z(){return{data:e.tableData.data,total:e.tableData.data.length}}function $(l,m){_.value=m.name,r("table-action",m)}function A(){r("refresh"),r("load-data",v.value),R.logger.info(C.TABLE_REFRESH_BUTTON_CLICKED)}function U(){v.value=e.pageOffset-e.pageSize,r("load-data",e.pageOffset-e.pageSize)}function W(){v.value=e.pageOffset+e.pageSize,r("load-data",e.pageOffset+e.pageSize)}function V({headerKey:l}){return{class:["warnings"].includes(l)?"text-center":["details"].includes(l)?"text-right":""}}function K({name:l}){const m=e.selectedEntityName||e.tableData.data[0].name;return{class:l===m?"is-selected":""}}return(l,m)=>{const h=G("router-link");return s(),u("div",le,[b("div",re,[E(l.$slots,"additionalControls",{},void 0,!0),o(),g(f(w),{class:"refresh-button",appearance:"primary",disabled:a.isLoading,icon:"redo","data-testid":"data-overview-refresh-button",onClick:A},{default:n(()=>[o(`
        Refresh
      `)]),_:1},8,["disabled"])]),o(),a.isLoading?(s(),c(Z,{key:0})):a.error!==null?(s(),c(J,{key:1,error:a.error},null,8,["error"])):a.isEmpty?(s(),c(x,{key:2})):(s(),u("div",de,[!a.tableDataIsEmpty&&a.tableData?(s(),u("div",ce,[(s(),c(f(X),{key:f(q),class:B(["data-overview-table",{"data-table-is-hidden":a.tableDataIsEmpty}]),fetcher:z,headers:f(N),"cell-attrs":V,"row-attrs":K,"disable-pagination":"","is-clickable":"","data-testid":"data-overview-table","onRow:click":$},O({status:n(({rowValue:t})=>[t?(s(),c(ee,{key:0,status:t},null,8,["status"])):(s(),u(y,{key:1},[o(`
              \u2014
            `)],64))]),tags:n(({rowValue:t})=>[g(te,{tags:t},null,8,["tags"])]),name:n(({row:t,rowValue:i})=>[t.nameRoute?(s(),c(h,{key:0,to:t.nameRoute},{default:n(()=>[o(d(i),1)]),_:2},1032,["to"])):(s(),u(y,{key:1},[o(d(i),1)],64))]),mesh:n(({row:t,rowValue:i})=>[t.meshRoute?(s(),c(h,{key:0,to:t.meshRoute},{default:n(()=>[o(d(i),1)]),_:2},1032,["to"])):(s(),u(y,{key:1},[o(d(i),1)],64))]),service:n(({row:t,rowValue:i})=>[t.serviceInsightRoute?(s(),c(h,{key:0,to:t.serviceInsightRoute},{default:n(()=>[o(d(i),1)]),_:2},1032,["to"])):(s(),u(y,{key:1},[o(d(i),1)],64))]),zone:n(({row:t,rowValue:i})=>[t.zoneRoute?(s(),c(h,{key:0,to:t.zoneRoute},{default:n(()=>[o(d(i),1)]),_:2},1032,["to"])):(s(),u(y,{key:1},[o(d(i),1)],64))]),totalUpdates:n(({row:t})=>[b("span",null,d(t.totalUpdates),1)]),dpVersion:n(({row:t,rowValue:i})=>[b("div",{class:B({"with-warnings":t.unsupportedEnvoyVersion||t.unsupportedKumaDPVersion||t.kumaDpAndKumaCpMismatch})},d(i),3)]),envoyVersion:n(({row:t,rowValue:i})=>[b("div",{class:B({"with-warnings":t.unsupportedEnvoyVersion})},d(i),3)]),_:2},[Q(f(T),t=>({name:t,fn:n(({rowValue:i,row:j})=>[E(l.$slots,t,{rowValue:i,row:j},void 0,!0)])})),a.showWarnings?{name:"warnings",fn:n(({row:t})=>[t.withWarnings?(s(),c(f(k),{key:0,class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20"})):p("",!0)]),key:"0"}:void 0,a.showDetails?{name:"details",fn:n(({row:t})=>[g(f(w),{class:"detail-link",appearance:"btn-link",to:t.nameRoute},{icon:n(()=>[g(f(k),{icon:t.warnings.length>0?"warning":"info",color:t.warnings.length>0?"var(--black-75)":"var(--blue-500)","secondary-color":t.warnings.length>0?"var(--yellow-300)":void 0,size:"16","hide-title":""},null,8,["icon","color","secondary-color"])]),default:n(()=>[o(`
              Details
            `)]),_:2},1032,["to"])]),key:"1"}:void 0]),1032,["class","headers"])),o(),g(oe,{"has-previous":v.value>0,"has-next":Boolean(a.next),onNext:W,onPrevious:U},null,8,["has-previous","has-next"])])):p("",!0),o(),a.tableDataIsEmpty&&a.tableData?(s(),c(x,{key:1},O({title:n(()=>[ue,o(),a.emptyState.title?(s(),u("p",fe,d(a.emptyState.title),1)):(s(),u("p",ge,`
            No items found
          `))]),_:2},[a.emptyState.message?{name:"message",fn:n(()=>[o(d(a.emptyState.message),1)]),key:"0"}:void 0]),1024)):p("",!0),o(),l.$slots.content?(s(),u("div",me,[E(l.$slots,"content",{},void 0,!0)])):p("",!0)]))])}}});const De=L(pe,[["__scopeId","data-v-ae85a567"]]);export{De as D,we as Q};
