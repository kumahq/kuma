import{k as y,cl as T,m as p,cm as w,cn as x,cc as h,o as s,c as o,b as n,e as a,bV as r,j as i,F as b,cd as I,i as R,w as f,a as C,bX as V,bY as K,co as B}from"./index.0a811bc4.js";const F={name:"SubscriptionDetails",components:{KAlert:T,KIcon:p},props:{details:{type:Object,required:!0},isDiscoverySubscription:{type:Boolean,default:!1}},computed:{detailsIterator(){var e;if(this.isDiscoverySubscription){const{lastUpdateTime:_,total:t,...d}=this.details.status;return d}return(e=this.details.status)==null?void 0:e.stat}},methods:{formatValue(e){return e?parseInt(e,10).toLocaleString("en").toString():0},readableDate(e){return w(e)},humanReadable(e){return x(e)},formatError(e){return e==="--"?"error calculating":e}}},l=e=>(V("data-v-ebd10bf9"),e=e(),K(),e),L={key:0},N=l(()=>a("h5",{class:"overview-tertiary-title"},`
        General Information:
      `,-1)),j={key:0},A=l(()=>a("strong",null,"Global Instance ID:",-1)),E={class:"mono"},P={key:1},q=l(()=>a("strong",null,"Control Plane Instance ID:",-1)),G={class:"mono"},H={key:2},O=l(()=>a("strong",null,"Last Connected:",-1)),M={key:3},U=l(()=>a("strong",null,"Last Disconnected:",-1)),W={key:1},X={class:"overview-stat-grid"},Y={class:"overview-tertiary-title"},z={class:"mono"};function J(e,_,t,d,g,c){const D=h("KIcon"),k=h("KAlert");return s(),o("div",null,[t.details.globalInstanceId||t.details.connectTime||t.details.disconnectTime?(s(),o("div",L,[N,n(),a("ul",null,[t.details.globalInstanceId?(s(),o("li",j,[A,n(`\xA0
          `),a("span",E,r(t.details.globalInstanceId),1)])):i("",!0),n(),t.details.controlPlaneInstanceId?(s(),o("li",P,[q,n(`\xA0
          `),a("span",G,r(t.details.controlPlaneInstanceId),1)])):i("",!0),n(),t.details.connectTime?(s(),o("li",H,[O,n(`\xA0
          `+r(c.readableDate(t.details.connectTime)),1)])):i("",!0),n(),t.details.disconnectTime?(s(),o("li",M,[U,n(`\xA0
          `+r(c.readableDate(t.details.disconnectTime)),1)])):i("",!0)])])):i("",!0),n(),c.detailsIterator?(s(),o("div",W,[a("ul",X,[(s(!0),o(b,null,I(c.detailsIterator,(S,u)=>(s(),o("li",{key:u},[a("h6",Y,r(c.humanReadable(u))+`:
          `,1),n(),a("ul",null,[(s(!0),o(b,null,I(S,(v,m)=>(s(),o("li",{key:m},[a("strong",null,r(c.humanReadable(m))+":",1),n(`\xA0
              `),a("span",z,r(c.formatError(c.formatValue(v))),1)]))),128))])]))),128))])])):(s(),R(k,{key:2,appearance:"info",class:"mt-4"},{alertIcon:f(()=>[C(D,{icon:"portal"})]),alertMessage:f(()=>[n(`
        There are no subscription statistics for `),a("strong",null,r(t.details.id),1)]),_:1}))])}const se=y(F,[["render",J],["__scopeId","data-v-ebd10bf9"]]),Q={name:"SubscriptionHeader",props:{details:{type:Object,required:!0}},methods:{rawReadableDateFilter(e){return B(e)}}},Z={class:"text-lg font-medium"},$={class:"color-green-500"},ee={key:0,class:"ml-4 color-red-600"};function te(e,_,t,d,g,c){return s(),o("h4",Z,[a("span",$,`
      Connect time: `+r(c.rawReadableDateFilter(t.details.connectTime)),1),n(),t.details.disconnectTime?(s(),o("span",ee,`
      Disconnect time: `+r(c.rawReadableDateFilter(t.details.disconnectTime)),1)):i("",!0)])}const ne=y(Q,[["render",te]]);export{se as S,ne as a};
