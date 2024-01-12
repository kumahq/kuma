import{K as E}from"./index-fce48c05.js";import{d as ee,L as te,B as V,a as u,o as n,b as v,w as e,e as o,W as F,m as D,f as s,t as m,l as k,ay as M,c as f,F as z,D as S,p as x,P as ne,az as oe,C as se,z as ae,_ as ie}from"./index-cf10d15e.js";import{A as le}from"./AppCollection-52f1b5ae.js";import{_ as re}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-0ce220f8.js";import{E as ce}from"./ErrorBlock-ce60392d.js";import{S as me}from"./StatusBadge-81868cb0.js";import{S as pe}from"./SummaryView-6eab098f.js";import{_ as ue}from"./WarningIcon.vue_vue_type_script_setup_true_lang-990b7d32.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-dae16a2a.js";import"./TextWithCopyButton-b8bd594c.js";import"./CopyButton-0aa5d830.js";const de=["data-testid"],_e=ee({__name:"ZoneListView",setup(fe){const O=te(),K=V(!1),N=V(!1),I=V(""),R=V({}),A=V({}),T=i=>{let d="offline";return i.length>0&&(d="online",typeof i[i.length-1].disconnectTime<"u"&&(d="offline")),d},P=i=>{const d="zoneIngress";R.value=i.items.reduce((p,_)=>{var y,w;const l=(y=_[d])==null?void 0:y.zone;if(typeof l<"u"){typeof p[l]>"u"&&(p[l]={online:[],offline:[]});const b=((w=_[`${d}Insight`])==null?void 0:w.subscriptions)||[],C=T(b);p[l][C].push(_)}return p},{})},W=i=>{const d="zoneEgress";A.value=i.items.reduce((p,_)=>{var y,w;const l=(y=_[d])==null?void 0:y.zone;if(typeof l<"u"){typeof p[l]>"u"&&(p[l]={online:[],offline:[]});const b=((w=_[`${d}Insight`])==null?void 0:w.subscriptions)||[],C=T(b);p[l][C].push(_)}return p},{})};async function j(){await O.deleteZone({name:I.value})}function Z(){K.value=!K.value}function G(i){Z(),I.value=i}function U(i){N.value=(i==null?void 0:i.items.length)>0}return(i,d)=>{const p=u("RouteTitle"),_=u("KButton"),l=u("DataSource"),y=u("RouterLink"),w=u("KTooltip"),b=u("KDropdownItem"),C=u("KDropdown"),H=u("KCard"),J=u("RouterView"),Q=u("AppView"),X=u("RouteView");return n(),v(l,{src:"/me"},{default:e(({data:q})=>[q?(n(),v(X,{key:0,name:"zone-cp-list-view",params:{page:1,size:q.pageSize,zone:""}},{default:e(({route:r,t:c,can:h})=>[o(Q,null,F({title:e(()=>[D("h1",null,[o(p,{title:c("zone-cps.routes.items.title")},null,8,["title"])])]),default:e(()=>[s(),s(),o(l,{src:`/zone-cps?page=${r.params.page}&size=${r.params.size}`,onChange:U},{default:e(({data:g,error:$,refresh:Y})=>[o(l,{src:"/zone-ingress-overviews?page=1&size=100",onChange:P}),s(),o(l,{src:"/zone-egress-overviews?page=1&size=100",onChange:W}),s(),o(H,null,{default:e(()=>[$!==void 0?(n(),v(ce,{key:0,error:$},null,8,["error"])):(n(),v(le,{key:1,class:"zone-cp-collection","data-testid":"zone-cp-collection",headers:[{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Type",key:"type"},{label:"Ingresses (online / total)",key:"ingress"},{label:"Egresses (online / total)",key:"egress"},{label:"Status",key:"state"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":r.params.page,"page-size":r.params.size,total:g==null?void 0:g.total,items:g==null?void 0:g.items,error:$,"empty-state-title":h("create zones")?c("zone-cps.empty_state.title"):void 0,"empty-state-message":h("create zones")?c("zone-cps.empty_state.message"):void 0,"empty-state-cta-to":h("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":h("create zones")?c("zones.index.create"):void 0,"is-selected-row":t=>t.name===r.params.zone,onChange:r.update},F({name:e(({row:t})=>[o(y,{to:{name:"zone-cp-detail-view",params:{zone:t.name},query:{page:r.params.page,size:r.params.size}}},{default:e(()=>[s(m(t.name),1)]),_:2},1032,["to"])]),zoneCpVersion:e(({row:t})=>[s(m(k(M)(t.zoneInsight,"version.kumaCp.version",c("common.collection.none"))),1)]),type:e(({row:t})=>[s(m(t.zoneInsight.environment.length>0?t.zoneInsight.environment:"kubernetes"),1)]),ingress:e(({row:t})=>[(n(!0),f(z,null,S([R.value[t.name]||{online:[],offline:[]}],a=>(n(),f(z,null,[s(m(a.online.length)+" / "+m(a.online.length+a.offline.length),1)],64))),256))]),egress:e(({row:t})=>[(n(!0),f(z,null,S([A.value[t.name]||{online:[],offline:[]}],a=>(n(),f(z,null,[s(m(a.online.length)+" / "+m(a.online.length+a.offline.length),1)],64))),256))]),state:e(({row:t})=>[o(me,{status:t.state},null,8,["status"])]),warnings:e(({row:t})=>[(n(!0),f(z,null,S([{version_mismatch:!k(M)(t.zoneInsight,"version.kumaCp.kumaCpGlobalCompatible","true"),store_memory:t.zoneInsight.store==="memory"}],a=>(n(),f(z,{key:`${a.version_mismatch}-${a.store_memory}`},[Object.values(a).some(B=>B)?(n(),v(w,{key:0},{content:e(()=>[D("ul",null,[(n(!0),f(z,null,S(a,(B,L)=>(n(),f(z,{key:L},[B?(n(),f("li",{key:0,"data-testid":`warning-${L}`},m(c(`zone-cps.list.${L}`)),9,de)):x("",!0)],64))),128))])]),default:e(()=>[s(),o(ue,{"data-testid":"warning",class:"mr-1",size:k(E),"hide-title":""},null,8,["size"])]),_:2},1024)):(n(),f(z,{key:1},[s(m(c("common.collection.none")),1)],64))],64))),128))]),details:e(({row:t})=>[o(y,{class:"details-link","data-testid":"details-link",to:{name:"zone-cp-detail-view",params:{zone:t.name}}},{default:e(()=>[s(m(c("common.collection.details_link"))+" ",1),o(k(ne),{display:"inline-block",decorative:"",size:k(E)},null,8,["size"])]),_:2},1032,["to"])]),_:2},[h("create zones")?{name:"actions",fn:e(({row:t})=>[o(C,{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[o(_,{class:"non-visual-button",appearance:"secondary","icon-only":""},{default:e(()=>[o(k(oe))]),_:1})]),items:e(()=>[o(b,{"has-divider":"",danger:"","data-testid":"dropdown-delete-item",onClick:a=>G(t.name)},{default:e(()=>[s(m(c("common.collection.actions.delete")),1)]),_:2},1032,["onClick"])]),_:2},1024)]),key:"0"}:void 0]),1032,["headers","page-number","page-size","total","items","error","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"]))]),_:2},1024),s(),r.params.zone?(n(),v(J,{key:0},{default:e(t=>[o(pe,{onClose:a=>r.replace({name:"zone-cp-list-view",query:{page:r.params.page,size:r.params.size}})},{default:e(()=>[(n(),v(se(t.Component),{name:r.params.zone,"zone-overview":g==null?void 0:g.items.find(a=>a.name===r.params.zone)},null,8,["name","zone-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):x("",!0),s(),K.value?(n(),v(re,{key:1,"confirmation-text":I.value,"delete-function":j,"is-visible":"","action-button-text":c("common.delete_modal.proceed_button"),title:c("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:Z,onDelete:()=>{Z(),Y()}},{"body-content":e(()=>[D("p",null,m(c("common.delete_modal.text1",{type:"Zone",name:I.value})),1),s(),D("p",null,m(c("common.delete_modal.text2")),1)]),_:2},1032,["confirmation-text","action-button-text","title","onDelete"])):x("",!0)]),_:2},1032,["src"])]),_:2},[h("create zones")&&N.value?{name:"actions",fn:e(()=>[o(_,{appearance:"primary",to:{name:"zone-create-view"},"data-testid":"create-zone-link"},{default:e(()=>[o(k(ae),{size:k(E)},null,8,["size"]),s(" "+m(c("zones.index.create")),1)]),_:2},1024)]),key:"0"}:void 0]),1024)]),_:2},1032,["params"])):x("",!0)]),_:1})}}});const De=ie(_e,[["__scopeId","data-v-97ded327"]]);export{De as default};
