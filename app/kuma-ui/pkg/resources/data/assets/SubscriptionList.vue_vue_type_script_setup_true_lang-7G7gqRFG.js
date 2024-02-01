import{d as f,l as S,I as _,o,c as l,b as y,w as h,al as V,q as r,e as w,t as e,f as t,am as T,r as x,p as v,m as n,F as k,E as $,_ as H,x as Z,y as A}from"./index-ANwvg_A1.js";import{A as U,a as D}from"./AccordionList-gNVeK0Wr.js";const O={class:"stack"},j={key:1},B={key:0,class:"intro"},P={class:"row"},N={class:"header"},q={class:"header"},E=["data-testid"],F={class:"type"},R=f({__name:"SubscriptionDetails",props:{subscription:{type:Object,required:!0},isDiscoverySubscription:{type:Boolean,default:!1}},setup(d){const{t:c}=S(),p=d,s=_(()=>{var u;let a;if("controlPlaneInstanceId"in p.subscription){const{lastUpdateTime:i,total:C,...m}=p.subscription.status;a=m}else a=((u=p.subscription.status)==null?void 0:u.stat)??{};return a?Object.entries(a).map(([i,C])=>{const{responsesSent:m="0",responsesAcknowledged:b="0",responsesRejected:g="0"}=C;return{type:i,responsesSent:m,responsesAcknowledged:b,responsesRejected:g}}):[]});return(a,u)=>(o(),l("div",O,[s.value.length===0?(o(),y(r(T),{key:0,appearance:"info"},{alertIcon:h(()=>[w(r(V))]),alertMessage:h(()=>[t(e(r(c)("common.detail.subscriptions.no_stats",{id:p.subscription.id})),1)]),_:1})):(o(),l("div",j,[a.$slots.default?(o(),l("div",B,[x(a.$slots,"default",{},void 0,!0)])):v("",!0),t(),n("div",P,[n("div",N,e(r(c)("common.detail.subscriptions.type")),1),t(),n("div",q,e(r(c)("common.detail.subscriptions.responses_sent_acknowledged")),1)]),t(),(o(!0),l(k,null,$(s.value,(i,C)=>(o(),l("div",{key:C,class:"row","data-testid":`subscription-status-${i.type}`},[n("div",F,e(r(c)(`http.api.property.${i.type}`)),1),t(),n("div",null,e(i.responsesSent)+"/"+e(i.responsesAcknowledged),1)],8,E))),128))]))]))}}),z=H(R,[["__scopeId","data-v-c3ee36ce"]]),G="data:image/svg+xml,%3csvg%20width='16'%20height='16'%20xmlns='http://www.w3.org/2000/svg'%3e%3cmask%20id='mask0_1449_15874'%20style='mask-type:alpha'%20maskUnits='userSpaceOnUse'%20x='0'%20y='0'%20width='16'%20height='16'%3e%3crect%20width='16'%20height='16'%20fill='%23d9d9d9'/%3e%3c/mask%3e%3cg%20mask='url(%23mask0_1449_15874)'%3e%3cpath%20d='M7.33333%2014.4833L2.66667%2011.7999C2.45556%2011.6777%202.29167%2011.5166%202.175%2011.3166C2.05833%2011.1166%202%2010.8944%202%2010.6499V5.34992C2%205.10547%202.05833%204.88325%202.175%204.68325C2.29167%204.48325%202.45556%204.32214%202.66667%204.19992L7.33333%201.51659C7.54444%201.39436%207.76667%201.33325%208%201.33325C8.23333%201.33325%208.45556%201.39436%208.66667%201.51659L13.3333%204.19992C13.5444%204.32214%2013.7083%204.48325%2013.825%204.68325C13.9417%204.88325%2014%205.10547%2014%205.34992V10.6499C14%2010.8944%2013.9417%2011.1166%2013.825%2011.3166C13.7083%2011.5166%2013.5444%2011.6777%2013.3333%2011.7999L8.66667%2014.4833C8.45556%2014.6055%208.23333%2014.6666%208%2014.6666C7.76667%2014.6666%207.54444%2014.6055%207.33333%2014.4833ZM7.33333%208.38325V12.9499L8%2013.3333L8.66667%2012.9499V8.38325L12.6667%206.06659V5.36659L11.95%204.94992L8%207.23325L4.05%204.94992L3.33333%205.36659V6.06659L7.33333%208.38325Z'%20fill='%23b6b6bd'/%3e%3c/g%3e%3c/svg%3e",J="data:image/svg+xml,%3csvg%20width='17'%20height='16'%20xmlns='http://www.w3.org/2000/svg'%3e%3cmask%20id='mask0_1449_15879'%20style='mask-type:alpha'%20maskUnits='userSpaceOnUse'%20x='0'%20y='0'%20width='17'%20height='16'%3e%3crect%20x='0.333252'%20width='16'%20height='16'%20fill='%23d9d9d9'/%3e%3c/mask%3e%3cg%20mask='url(%23mask0_1449_15879)'%3e%3cpath%20d='M12.9998%2014C12.5665%2014%2012.1776%2013.875%2011.8332%2013.625C11.4887%2013.375%2011.2498%2013.0556%2011.1165%2012.6667H7.6665C6.93317%2012.6667%206.30539%2012.4056%205.78317%2011.8833C5.26095%2011.3611%204.99984%2010.7333%204.99984%2010C4.99984%209.26667%205.26095%208.63889%205.78317%208.11667C6.30539%207.59444%206.93317%207.33333%207.6665%207.33333H8.99984C9.3665%207.33333%209.68039%207.20278%209.9415%206.94167C10.2026%206.68056%2010.3332%206.36667%2010.3332%206C10.3332%205.63333%2010.2026%205.31944%209.9415%205.05833C9.68039%204.79722%209.3665%204.66667%208.99984%204.66667H5.54984C5.40539%205.05556%205.16373%205.375%204.82484%205.625C4.48595%205.875%204.09984%206%203.6665%206C3.11095%206%202.63873%205.80556%202.24984%205.41667C1.86095%205.02778%201.6665%204.55556%201.6665%204C1.6665%203.44444%201.86095%202.97222%202.24984%202.58333C2.63873%202.19444%203.11095%202%203.6665%202C4.09984%202%204.48595%202.125%204.82484%202.375C5.16373%202.625%205.40539%202.94444%205.54984%203.33333H8.99984C9.73317%203.33333%2010.3609%203.59444%2010.8832%204.11667C11.4054%204.63889%2011.6665%205.26667%2011.6665%206C11.6665%206.73333%2011.4054%207.36111%2010.8832%207.88333C10.3609%208.40556%209.73317%208.66667%208.99984%208.66667H7.6665C7.29984%208.66667%206.98595%208.79722%206.72484%209.05833C6.46373%209.31945%206.33317%209.63333%206.33317%2010C6.33317%2010.3667%206.46373%2010.6806%206.72484%2010.9417C6.98595%2011.2028%207.29984%2011.3333%207.6665%2011.3333H11.1165C11.2609%2010.9444%2011.5026%2010.625%2011.8415%2010.375C12.1804%2010.125%2012.5665%2010%2012.9998%2010C13.5554%2010%2014.0276%2010.1944%2014.4165%2010.5833C14.8054%2010.9722%2014.9998%2011.4444%2014.9998%2012C14.9998%2012.5556%2014.8054%2013.0278%2014.4165%2013.4167C14.0276%2013.8056%2013.5554%2014%2012.9998%2014ZM3.6665%204.66667C3.85539%204.66667%204.01373%204.60278%204.1415%204.475C4.26928%204.34722%204.33317%204.18889%204.33317%204C4.33317%203.81111%204.26928%203.65278%204.1415%203.525C4.01373%203.39722%203.85539%203.33333%203.6665%203.33333C3.47761%203.33333%203.31928%203.39722%203.1915%203.525C3.06373%203.65278%202.99984%203.81111%202.99984%204C2.99984%204.18889%203.06373%204.34722%203.1915%204.475C3.31928%204.60278%203.47761%204.66667%203.6665%204.66667Z'%20fill='%2307a88d'/%3e%3c/g%3e%3c/svg%3e",K="data:image/svg+xml,%3csvg%20width='17'%20height='16'%20xmlns='http://www.w3.org/2000/svg'%3e%3cmask%20id='mask0_1449_15884'%20style='mask-type:alpha'%20maskUnits='userSpaceOnUse'%20x='0'%20y='0'%20width='17'%20height='16'%3e%3crect%20x='0.666504'%20width='16'%20height='16'%20fill='%23d9d9d9'/%3e%3c/mask%3e%3cg%20mask='url(%23mask0_1449_15884)'%3e%3cpath%20d='M14.3001%2015.5335L1.1001%202.33354L2.0501%201.38354L15.2501%2014.5835L14.3001%2015.5335ZM8.0001%2012.6669C7.26676%2012.6669%206.63899%2012.4058%206.11676%2011.8835C5.59454%2011.3613%205.33343%2010.7335%205.33343%2010.0002C5.33343%209.26688%205.59454%208.6391%206.11676%208.11688C6.63899%207.59466%207.26676%207.33355%208.0001%207.33355L9.33343%208.66688H8.0001C7.63343%208.66688%207.31954%208.79743%207.05843%209.05855C6.79732%209.31966%206.66676%209.63354%206.66676%2010.0002C6.66676%2010.3669%206.79732%2010.6808%207.05843%2010.9419C7.31954%2011.203%207.63343%2011.3335%208.0001%2011.3335H12.0001L14.3668%2013.7002C14.2112%2013.7891%2014.0473%2013.8613%2013.8751%2013.9169C13.7029%2013.9724%2013.5223%2014.0002%2013.3334%2014.0002C12.9001%2014.0002%2012.5112%2013.8752%2012.1668%2013.6252C11.8223%2013.3752%2011.5834%2013.0558%2011.4501%2012.6669H8.0001ZM15.2168%2012.6502L12.6834%2010.1169C12.7834%2010.0835%2012.8862%2010.0558%2012.9918%2010.0335C13.0973%2010.0113%2013.2112%2010.0002%2013.3334%2010.0002C13.889%2010.0002%2014.3612%2010.1947%2014.7501%2010.5835C15.139%2010.9724%2015.3334%2011.4447%2015.3334%2012.0002C15.3334%2012.1224%2015.3223%2012.2363%2015.3001%2012.3419C15.2779%2012.4474%2015.2501%2012.5502%2015.2168%2012.6502ZM10.8001%208.23355L9.81676%207.25021C10.0723%207.15021%2010.2779%206.9891%2010.4334%206.76688C10.589%206.54466%2010.6668%206.2891%2010.6668%206.00021C10.6668%205.63355%2010.5362%205.31966%2010.2751%205.05855C10.014%204.79743%209.7001%204.66688%209.33343%204.66688H7.23343L5.9001%203.33354H9.33343C10.0668%203.33354%2010.6945%203.59466%2011.2168%204.11688C11.739%204.6391%2012.0001%205.26688%2012.0001%206.00021C12.0001%206.46688%2011.889%206.89466%2011.6668%207.28355C11.4445%207.67243%2011.1557%207.9891%2010.8001%208.23355ZM4.0001%206.00021C3.44454%206.00021%202.97232%205.80577%202.58343%205.41688C2.19454%205.02799%202.0001%204.55577%202.0001%204.00021C2.0001%203.64466%202.08899%203.31688%202.26676%203.01688C2.44454%202.71688%202.67788%202.47799%202.96676%202.30021L5.7001%205.03355C5.52232%205.32243%205.28343%205.55577%204.98343%205.73355C4.68343%205.91132%204.35565%206.00021%204.0001%206.00021Z'%20fill='%23b6b6bd'/%3e%3c/g%3e%3c/svg%3e",I=d=>(Z("data-v-991b71e7"),d=d(),A(),d),Q={class:"subscription-header"},W={class:"instance-id"},X=I(()=>n("img",{src:G},null,-1)),Y=I(()=>n("img",{src:J},null,-1)),s0={key:0},e0=I(()=>n("img",{src:K},null,-1)),t0={class:"responses-sent-acknowledged"},n0=f({__name:"SubscriptionHeader",props:{subscription:{type:Object,required:!0}},setup(d){const{t:c,formatIsoDate:p}=S(),s=d,a=_(()=>"globalInstanceId"in s.subscription?s.subscription.globalInstanceId:null),u=_(()=>"controlPlaneInstanceId"in s.subscription?s.subscription.controlPlaneInstanceId:null),i=_(()=>s.subscription.connectTime?p(s.subscription.connectTime):null),C=_(()=>s.subscription.disconnectTime?p(s.subscription.disconnectTime):null),m=_(()=>{var L;const{responsesSent:b=0,responsesAcknowledged:g=0,responsesRejected:M=0}=((L=s.subscription.status)==null?void 0:L.total)??{};return{responsesSent:b,responsesAcknowledged:g,responsesRejected:M}});return(b,g)=>(o(),l("header",Q,[n("span",W,[X,t(),a.value?(o(),l(k,{key:0},[n("b",null,e(r(c)("http.api.property.globalInstanceId")),1),t(": "+e(a.value),1)],64)):u.value?(o(),l(k,{key:1},[n("b",null,e(r(c)("http.api.property.controlPlaneInstanceId")),1),t(": "+e(u.value),1)],64)):v("",!0)]),t(),n("span",null,[Y,t(),n("b",null,e(r(c)("common.detail.subscriptions.connect_time")),1),t(": "+e(i.value),1)]),t(),C.value?(o(),l("span",s0,[e0,t(),n("b",null,e(r(c)("common.detail.subscriptions.disconnect_time")),1),t(": "+e(C.value),1)])):v("",!0),t(),n("span",t0,e(r(c)("common.detail.subscriptions.responses_sent_acknowledged"))+`:

      `+e(m.value.responsesSent)+"/"+e(m.value.responsesAcknowledged),1)]))}}),o0=H(n0,[["__scopeId","data-v-991b71e7"]]),i0=f({__name:"SubscriptionList",props:{subscriptions:{}},setup(d){const c=d,p=_(()=>{const s=Array.from(c.subscriptions);return s.reverse(),s});return(s,a)=>(o(),y(D,null,{default:h(()=>[(o(!0),l(k,null,$(p.value,(u,i)=>(o(),y(U,{key:i},{"accordion-header":h(()=>[w(o0,{subscription:u},null,8,["subscription"])]),"accordion-content":h(()=>[w(z,{subscription:u},{default:h(()=>[s.$slots.default?x(s.$slots,"default",{key:0}):v("",!0)]),_:2},1032,["subscription"])]),_:2},1024))),128))]),_:3}))}});export{i0 as _};
