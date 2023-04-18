import{d as D,i as w,o as t,c as n,e as a,f as s,z as i,y as u,u as r,L as g,F as I,A as k,j as x,w as b,a as R,b as j,X as C,N as B,O as L,J as N,a6 as S}from"./index-5f1fbf13.js";const p=e=>(B("data-v-4f37a467"),e=e(),L(),e),O={key:0},V=p(()=>s("h5",{class:"overview-tertiary-title"},`
        General Information:
      `,-1)),P={key:0},$=p(()=>s("strong",null,"Global Instance ID:",-1)),A={key:1},q=p(()=>s("strong",null,"Control Plane Instance ID:",-1)),E={key:2},F=p(()=>s("strong",null,"Last Connected:",-1)),G={key:3},z=p(()=>s("strong",null,"Last Disconnected:",-1)),H={key:1},J={class:"columns",style:{"--columns":"4"}},M={key:0},U={class:"overview-tertiary-title"},X=D({__name:"SubscriptionDetails",props:{details:{type:Object,required:!0},isDiscoverySubscription:{type:Boolean,default:!1}},setup(e){const l=e,h={responsesSent:"Responses Sent",responsesAcknowledged:"Responses Acknowledged",responsesRejected:"Responses Rejected"},f=w(()=>{var y,_;let o;if(l.isDiscoverySubscription){const{lastUpdateTime:m,total:d,...c}=l.details.status;o=c}(y=l.details.status)!=null&&y.stat&&(o=(_=l.details.status)==null?void 0:_.stat);for(const m in o){const d=o[m];for(const c in d)c in h&&(d[h[c]]=d[c],delete d[c])}return o});function T(o){return o?parseInt(o,10).toLocaleString("en").toString():"0"}function v(o){return o==="--"?"error calculating":o}return(o,y)=>(t(),n("div",null,[e.details.globalInstanceId||e.details.connectTime||e.details.disconnectTime?(t(),n("div",O,[V,a(),s("ul",null,[e.details.globalInstanceId?(t(),n("li",P,[$,a(),s("span",null,i(e.details.globalInstanceId),1)])):u("",!0),a(),e.details.controlPlaneInstanceId?(t(),n("li",A,[q,a(),s("span",null,i(e.details.controlPlaneInstanceId),1)])):u("",!0),a(),e.details.connectTime?(t(),n("li",E,[F,a(" "+i(r(g)(e.details.connectTime)),1)])):u("",!0),a(),e.details.disconnectTime?(t(),n("li",G,[z,a(" "+i(r(g)(e.details.disconnectTime)),1)])):u("",!0)])])):u("",!0),a(),r(f)?(t(),n("div",H,[s("ul",J,[(t(!0),n(I,null,k(r(f),(_,m)=>(t(),n(I,{key:m},[Object.keys(_).length>0?(t(),n("li",M,[s("h6",U,i(m)+`:
            `,1),a(),s("ul",null,[(t(!0),n(I,null,k(_,(d,c)=>(t(),n("li",{key:c},[s("strong",null,i(c)+":",1),a(),s("span",null,i(v(T(d))),1)]))),128))])])):u("",!0)],64))),128))])])):(t(),x(r(C),{key:2,appearance:"info",class:"mt-4"},{alertIcon:b(()=>[R(r(j),{icon:"portal"})]),alertMessage:b(()=>[a(`
        There are no subscription statistics for `),s("strong",null,i(e.details.id),1)]),_:1}))]))}});const Z=N(X,[["__scopeId","data-v-4f37a467"]]),K={class:"text-lg font-medium"},Q={class:"color-green-500"},W={key:0,class:"ml-4 color-red-600"},ee=D({__name:"SubscriptionHeader",props:{details:{type:Object,required:!0}},setup(e){const l=e;return(h,f)=>(t(),n("h4",K,[s("span",Q,`
      Connect time: `+i(r(S)(l.details.connectTime)),1),a(),l.details.disconnectTime?(t(),n("span",W,`
      Disconnect time: `+i(r(S)(l.details.disconnectTime)),1)):u("",!0)]))}});export{Z as S,ee as _};
