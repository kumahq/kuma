import{m as w,a as x}from"./kongponents.es-3df60cd6.js";import{i as I,r as y}from"./store-f89d97b2.js";import{d as v,c as R,o as s,h as o,f as n,g as t,t as i,b as u,u as r,F as b,v as k,a as C,w as S,e as B,p as V,m as $}from"./runtime-dom.esm-bundler-91b41870.js";import{_ as j}from"./_plugin-vue_export-helper-c27b6911.js";const _=e=>(V("data-v-f81d90f6"),e=e(),$(),e),L={key:0},N=_(()=>t("h5",{class:"overview-tertiary-title"},`
        General Information:
      `,-1)),P={key:0},q=_(()=>t("strong",null,"Global Instance ID:",-1)),A={key:1},E=_(()=>t("strong",null,"Control Plane Instance ID:",-1)),F={key:2},G=_(()=>t("strong",null,"Last Connected:",-1)),O={key:3},H=_(()=>t("strong",null,"Last Disconnected:",-1)),M={key:1},U={class:"overview-stat-grid"},z={class:"overview-tertiary-title"},J=v({__name:"SubscriptionDetails",props:{details:{type:Object,required:!0},isDiscoverySubscription:{type:Boolean,default:!1}},setup(e){const l=e,f={responsesSent:"Responses Sent",responsesAcknowledged:"Responses Acknowledged",responsesRejected:"Responses Rejected"},h=R(()=>{var g,p;let a;if(l.isDiscoverySubscription){const{lastUpdateTime:m,total:d,...c}=l.details.status;a=c}(g=l.details.status)!=null&&g.stat&&(a=(p=l.details.status)==null?void 0:p.stat);for(const m in a){const d=a[m];for(const c in d)c in f&&(d[f[c]]=d[c],delete d[c])}return a});function D(a){return a?parseInt(a,10).toLocaleString("en").toString():"0"}function T(a){return a==="--"?"error calculating":a}return(a,g)=>(s(),o("div",null,[e.details.globalInstanceId||e.details.connectTime||e.details.disconnectTime?(s(),o("div",L,[N,n(),t("ul",null,[e.details.globalInstanceId?(s(),o("li",P,[q,n(),t("span",null,i(e.details.globalInstanceId),1)])):u("",!0),n(),e.details.controlPlaneInstanceId?(s(),o("li",A,[E,n(),t("span",null,i(e.details.controlPlaneInstanceId),1)])):u("",!0),n(),e.details.connectTime?(s(),o("li",F,[G,n(" "+i(r(I)(e.details.connectTime)),1)])):u("",!0),n(),e.details.disconnectTime?(s(),o("li",O,[H,n(" "+i(r(I)(e.details.disconnectTime)),1)])):u("",!0)])])):u("",!0),n(),r(h)?(s(),o("div",M,[t("ul",U,[(s(!0),o(b,null,k(r(h),(p,m)=>(s(),o("li",{key:m},[t("h6",z,i(m)+`:
          `,1),n(),t("ul",null,[(s(!0),o(b,null,k(p,(d,c)=>(s(),o("li",{key:c},[t("strong",null,i(c)+":",1),n(),t("span",null,i(T(D(d))),1)]))),128))])]))),128))])])):(s(),C(r(x),{key:2,appearance:"info",class:"mt-4"},{alertIcon:S(()=>[B(r(w),{icon:"portal"})]),alertMessage:S(()=>[n(`
        There are no subscription statistics for `),t("strong",null,i(e.details.id),1)]),_:1}))]))}});const te=j(J,[["__scopeId","data-v-f81d90f6"]]),K={class:"text-lg font-medium"},Q={class:"color-green-500"},W={key:0,class:"ml-4 color-red-600"},se=v({__name:"SubscriptionHeader",props:{details:{type:Object,required:!0}},setup(e){const l=e;return(f,h)=>(s(),o("h4",K,[t("span",Q,`
      Connect time: `+i(r(y)(l.details.connectTime)),1),n(),l.details.disconnectTime?(s(),o("span",W,`
      Disconnect time: `+i(r(y)(l.details.disconnectTime)),1)):u("",!0)]))}});export{te as S,se as _};
