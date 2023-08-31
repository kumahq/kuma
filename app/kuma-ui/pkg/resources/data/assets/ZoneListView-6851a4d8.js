import{o as D,A as F,i as q,E as K,q as U,S as W,K as A,r as P,_ as G,F as J,f as H}from"./RouteView.vue_vue_type_script_setup_true_lang-9ff59c45.js";import{d as L,l as N,o as a,a as i,w as e,C as Z,g as n,b as s,N as X,f as T,ai as Y,r as j,h as l,E as Q,i as h,k as ee,t as r,e as O,F as M,M as te,R as oe,H as S,x as ne,J as I}from"./index-6c44d272.js";import{_ as ae}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-090c3f11.js";import{_ as se}from"./RouteTitle.vue_vue_type_script_setup_true_lang-ce3600d0.js";import{_ as ie}from"./WarningIcon.vue_vue_type_script_setup_true_lang-7307ec83.js";const le=L({__name:"DeleteResourceModal",props:{actionButtonText:{type:String,required:!1,default:"Yes, delete"},confirmationText:{type:String,required:!1,default:""},deleteFunction:{type:Function,required:!0},isVisible:{type:Boolean,required:!0},modalId:{type:String,required:!0},title:{type:String,required:!1,default:"Delete"}},emits:["cancel","delete"],setup(w,{emit:p}){const c=w,d=N(!1);async function v(){d.value=!1;try{await c.deleteFunction(),p("delete")}catch{d.value=!0}}return(u,g)=>(a(),i(s(Y),{"action-button-text":c.actionButtonText,"confirmation-text":c.confirmationText,"is-visible":c.isVisible,"modal-id":c.modalId,title:c.title,type:"danger","data-testid":"delete-resource-modal",onCanceled:g[0]||(g[0]=R=>p("cancel")),onProceed:v},{"body-content":e(()=>[Z(u.$slots,"body-content"),n(),d.value?(a(),i(s(X),{key:0,class:"mt-4",appearance:"danger","is-dismissible":""},{alertMessage:e(()=>[Z(u.$slots,"error")]),_:3})):T("",!0)]),_:3},8,["action-button-text","confirmation-text","is-visible","modal-id","title"]))}}),re=L({__name:"ZoneListView",props:{page:{type:Number,required:!0},size:{type:Number,required:!0}},setup(w){const p=w,c=D(),d=N(!1),v=N(!1),u=N("");function g(m){return m.map(z=>{var C;const{name:k}=z,V={name:"zone-cp-detail-view",params:{zone:k}};let t="",f="kubernetes",_=!0;(((C=z.zoneInsight)==null?void 0:C.subscriptions)??[]).forEach(o=>{if(o.version&&o.version.kumaCp){t=o.version.kumaCp.version;const{kumaCpGlobalCompatible:x=!0}=o.version.kumaCp;_=x}o.config&&(f=JSON.parse(o.config).environment)});const b=J(z);return{detailViewRoute:V,name:k,status:b,zoneCpVersion:t,type:f,warnings:!_}})}async function R(){await c.deleteZone({name:u.value})}function E(){d.value=!d.value}function B(m){E(),u.value=m}function $(m){v.value=(m==null?void 0:m.items.length)>0}return(m,z)=>{const k=j("RouterLink");return a(),i(G,{name:"zone-cp-list-view"},{default:e(({route:V,t,can:f,env:_})=>[l(F,null,Q({title:e(()=>[h("h1",null,[l(se,{title:t("zone-cps.routes.items.title"),render:!0},null,8,["title"])])]),default:e(()=>[n(),n(),f("use zones")?(a(),i(q,{key:1,src:`/zone-cps?size=${p.size}&page=${p.page}`,onChange:$},{default:e(({data:y,error:b,refresh:C})=>[l(s(ee),null,{body:e(()=>[b!==void 0?(a(),i(K,{key:0,error:b},null,8,["error"])):(a(),i(U,{key:1,class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Type",key:"type"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":p.page,"page-size":p.size,total:y==null?void 0:y.total,items:y?g(y.items):void 0,error:b,"empty-state-title":_("KUMA_ZONE_CREATION_FLOW")==="enabled"?t("zone-cps.empty_state.title"):void 0,"empty-state-message":_("KUMA_ZONE_CREATION_FLOW")==="enabled"?t("zone-cps.empty_state.message"):void 0,"empty-state-cta-to":_("KUMA_ZONE_CREATION_FLOW")==="enabled"?{name:"zone-create-view"}:void 0,"empty-state-cta-text":_("KUMA_ZONE_CREATION_FLOW")==="enabled"?t("zones.index.create"):void 0,onChange:V.update},{name:e(({row:o,rowValue:x})=>[l(k,{to:o.detailViewRoute,"data-testid":"detail-view-link"},{default:e(()=>[n(r(x),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({rowValue:o})=>[n(r(o||t("common.collection.none")),1)]),type:e(({rowValue:o})=>[n(r(o||t("common.collection.none")),1)]),status:e(({rowValue:o})=>[o?(a(),i(W,{key:0,status:o},null,8,["status"])):(a(),O(M,{key:1},[n(r(t("common.collection.none")),1)],64))]),warnings:e(({rowValue:o})=>[o?(a(),i(s(te),{key:0,label:t("zone-cps.list.version_mismatch")},{default:e(()=>[l(ie,{class:"mr-1",size:s(A),"hide-title":""},null,8,["size"])]),_:2},1032,["label"])):(a(),O(M,{key:1},[n(r(t("common.collection.none")),1)],64))]),actions:e(({row:o})=>[l(s(oe),{class:"actions-dropdown","data-testid":"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[l(s(S),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:e(()=>[l(s(ne),{color:s(P),icon:"more",size:s(A)},null,8,["color","size"])]),_:1})]),items:e(()=>[l(s(I),{item:{to:o.detailViewRoute,label:t("common.collection.actions.view")}},null,8,["item"]),n(),f("create zones")?(a(),i(s(I),{key:0,"has-divider":"","is-dangerous":"","data-testid":"dropdown-delete-item",onClick:x=>B(o.name)},{default:e(()=>[n(r(t("common.collection.actions.delete")),1)]),_:2},1032,["onClick"])):T("",!0)]),_:2},1024)]),_:2},1032,["page-number","page-size","total","items","error","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]))]),_:2},1024),n(),d.value?(a(),i(le,{key:0,"confirmation-text":u.value,"delete-function":R,"is-visible":d.value,"modal-id":"delete-zone-modal","action-button-text":t("zones.delete.confirmModal.proceedText"),title:t("zones.delete.confirmModal.title"),onCancel:E,onDelete:()=>{E(),C()}},{"body-content":e(()=>[h("p",null,r(t("zones.delete.confirmModal.text1",{zoneName:u.value})),1),n(),h("p",null,r(t("zones.delete.confirmModal.text2")),1)]),error:e(()=>[n(r(t("zones.delete.confirmModal.errorText")),1)]),_:2},1032,["confirmation-text","is-visible","action-button-text","title","onDelete"])):T("",!0)]),_:2},1032,["src"])):(a(),i(ae,{key:0}))]),_:2},[f("create zones")&&v.value?{name:"actions",fn:e(()=>[l(s(S),{appearance:"primary",icon:"plus",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[n(r(t("zones.index.create")),1)]),_:2},1024)]),key:"0"}:void 0]),1024)]),_:1})}}});const _e=H(re,[["__scopeId","data-v-d0cdd1d2"]]);export{_e as default};
