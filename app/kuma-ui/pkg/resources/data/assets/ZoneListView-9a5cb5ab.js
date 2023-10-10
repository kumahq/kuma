import{d as j,P as F,v as I,r as z,o,g as r,w as e,h as l,a7 as M,m as V,l as a,i,Y as U,E as W,Z as q,D as m,U as H,j as k,F as h,a9 as J,G as O,k as x,s as Y,K as N,$ as Q,W as B,a0 as X,a1 as L,aG as ee,aH as te,q as oe}from"./index-ecc7df9d.js";import{_ as ne}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-9d3183b0.js";import{_ as ae}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-7b94cf9c.js";const se=["data-testid"],ie=j({__name:"ZoneListView",setup(le){const A=F(),Z=I(!1),R=I(!1),v=I("");function E(p){return p.map(g=>{var c;const{name:b}=g,S={name:"zone-cp-detail-view",params:{zone:b}};let y="",w="kubernetes",C=!1,f=!0;(((c=g.zoneInsight)==null?void 0:c.subscriptions)??[]).forEach(s=>{if(s.version&&s.version.kumaCp){y=s.version.kumaCp.version;const{kumaCpGlobalCompatible:u=!0}=s.version.kumaCp;f=u}if(s.config){const u=JSON.parse(s.config);w=u.environment,C=u.store.type==="memory"}});const t=te(g);return{detailViewRoute:S,name:b,status:t,zoneCpVersion:y,type:w,warnings:{version_mismatch:!f,store_memory:C}}})}async function T(){await A.deleteZone({name:v.value})}function D(){Z.value=!Z.value}function G(p){D(),v.value=p}function K(p){R.value=(p==null?void 0:p.items.length)>0}return(p,g)=>{const b=z("RouteTitle"),S=z("RouterLink"),y=z("DataSource"),w=z("AppView"),C=z("RouteView");return o(),r(y,{src:"/me"},{default:e(({data:f})=>[f?(o(),r(C,{key:0,name:"zone-cp-list-view",params:{page:1,size:f.pageSize}},{default:e(({route:d,t,can:c})=>[l(w,null,M({title:e(()=>[V("h1",null,[l(b,{title:t("zone-cps.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[a(),a(),c("use zones")?(o(),r(y,{key:1,src:`/zone-cps?page=${d.params.page}&size=${d.params.size}`,onChange:K},{default:e(({data:s,error:u,refresh:P})=>[l(i(U),null,{body:e(()=>[u!==void 0?(o(),r(W,{key:0,error:u},null,8,["error"])):(o(),r(q,{key:1,class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Type",key:"type"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":parseInt(d.params.page),"page-size":parseInt(d.params.size),total:s==null?void 0:s.total,items:s?E(s.items):void 0,error:u,"empty-state-title":c("create zones")?t("zone-cps.empty_state.title"):void 0,"empty-state-message":c("create zones")?t("zone-cps.empty_state.message"):void 0,"empty-state-cta-to":c("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":c("create zones")?t("zones.index.create"):void 0,onChange:d.update},{name:e(({row:n,rowValue:_})=>[l(S,{to:n.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[a(m(_),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({rowValue:n})=>[a(m(n||t("common.collection.none")),1)]),type:e(({rowValue:n})=>[a(m(n||t("common.collection.none")),1)]),status:e(({rowValue:n})=>[n?(o(),r(H,{key:0,status:n},null,8,["status"])):(o(),k(h,{key:1},[a(m(t("common.collection.none")),1)],64))]),warnings:e(({row:n})=>[Object.values(n.warnings).some(_=>_)?(o(),r(i(J),{key:0},{content:e(()=>[V("ul",null,[(o(!0),k(h,null,O(n.warnings,(_,$)=>(o(),k(h,{key:$},[_?(o(),k("li",{key:0,"data-testid":`warning-${$}`},m(t(`zone-cps.list.${$}`)),9,se)):x("",!0)],64))),128))])]),default:e(()=>[a(),l(Y,{"data-testid":"warning",class:"mr-1",size:i(N),"hide-title":""},null,8,["size"])]),_:2},1024)):(o(),k(h,{key:1},[a(m(t("common.collection.none")),1)],64))]),actions:e(({row:n})=>[l(i(Q),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[l(i(B),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[l(i(X),{size:i(N)},null,8,["size"])]),_:1})]),items:e(()=>[l(i(L),{item:{to:n.detailViewRoute,label:t("common.collection.actions.view")}},null,8,["item"]),a(),c("create zones")?(o(),r(i(L),{key:0,"has-divider":"","is-dangerous":"","data-testid":"dropdown-delete-item",onClick:_=>G(n.name)},{default:e(()=>[a(m(t("common.collection.actions.delete")),1)]),_:2},1032,["onClick"])):x("",!0)]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024),a(),Z.value?(o(),r(ae,{key:0,"confirmation-text":v.value,"delete-function":T,"is-visible":"","action-button-text":t("common.delete_modal.proceed_button"),title:t("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:D,onDelete:()=>{D(),P()}},{"body-content":e(()=>[V("p",null,m(t("common.delete_modal.text1",{type:"Zone",name:v.value})),1),a(),V("p",null,m(t("common.delete_modal.text2")),1)]),_:2},1032,["confirmation-text","action-button-text","title","onDelete"])):x("",!0)]),_:2},1032,["src"])):(o(),r(ne,{key:0}))]),_:2},[c("create zones")&&R.value?{name:"actions",fn:e(()=>[l(i(B),{appearance:"primary",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[l(i(ee),{size:i(N)},null,8,["size"]),a(" "+m(t("zones.index.create")),1)]),_:2},1024)]),key:"0"}:void 0]),1024)]),_:2},1032,["params"])):x("",!0)]),_:1})}}});const pe=oe(ie,[["__scopeId","data-v-4361cde8"]]);export{pe as default};
