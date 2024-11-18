import{d as K,I as j,F as A,o as s,m as b,w as i,c as p,G as ot,E as et,q as X,e as B,k as o,r as H,b as a,j as lt,p as V,a as u,t as r,l as m,H as _,J as q,n as ut,D as pt,Q as J,P as L,S as ct,a0 as mt,K as Q,$ as ft,a1 as yt,A as gt}from"./index-C_eW3RRu.js";import{q as vt}from"./kong-icons.es678-BrfIDRc7.js";import{S as _t}from"./SummaryView-C9wRmLik.js";import{T as rt}from"./TagList-B5j3u3SI.js";const bt=["aria-hidden"],at='<path d="M15 18L21 12L15 6L13.6 7.4L17.2 11H3V13H17.2L13.6 16.6L15 18Z" fill="currentColor"/>',kt=K({__name:"ForwardIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:j,validator:d=>{if(typeof d=="number"&&d>0)return!0;if(typeof d=="string"){const n=String(d).replace(/px/gi,""),e=Number(n);if(e&&!isNaN(e)&&Number.isInteger(e)&&e>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(d){const n=d,e=A(()=>{if(typeof n.size=="number"&&n.size>0)return`${n.size}px`;if(typeof n.size=="string"){const g=String(n.size).replace(/px/gi,""),c=Number(g);if(c&&!isNaN(c)&&Number.isInteger(c)&&c>0)return`${c}px`}return j}),x=A(()=>({boxSizing:"border-box",color:n.color,display:n.display,flexShrink:"0",height:e.value,lineHeight:"0",width:e.value,pointerEvents:n.decorative?"none":void 0})),N=g=>{const c={},C=Math.random().toString(36).substring(2,12);return g.replace(/id="([^"]+)"/g,(z,$)=>{const E=`${C}-${$}`;return c[$]=E,`id="${E}"`}).replace(/#([^\s^")]+)/g,(z,$)=>c[$]?`#${c[$]}`:z)},t=g=>{const c={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;","`":"&#039;"};return g.replace(/[<>"'`]/g,C=>c[C])},k=`${n.title?`<title data-testid="kui-icon-svg-title">${t(n.title)}</title>`:""}${n.staticIds?at:N(at)}`;return(g,c)=>(s(),b(et(d.as),{"aria-hidden":d.decorative?"true":void 0,class:"kui-icon forward-icon","data-testid":"kui-icon-wrapper-forward-icon",style:ot(x.value),tabindex:d.decorative?"-1":void 0},{default:i(()=>[(s(),p("svg",{"aria-hidden":d.decorative?"true":void 0,"data-testid":"kui-icon-svg-forward-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:k},null,8,bt))]),_:1},8,["aria-hidden","style","tabindex"]))}}),xt=["aria-hidden"],nt='<path d="M12 21V19H19V5H12V3H19C19.55 3 20.0208 3.19583 20.4125 3.5875C20.8042 3.97917 21 4.45 21 5V19C21 19.55 20.8042 20.0208 20.4125 20.4125C20.0208 20.8042 19.55 21 19 21H12ZM10 17L8.625 15.55L11.175 13H3V11H11.175L8.625 8.45L10 7L15 12L10 17Z" fill="currentColor"/>',$t=K({__name:"GatewayIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:j,validator:d=>{if(typeof d=="number"&&d>0)return!0;if(typeof d=="string"){const n=String(d).replace(/px/gi,""),e=Number(n);if(e&&!isNaN(e)&&Number.isInteger(e)&&e>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(d){const n=d,e=A(()=>{if(typeof n.size=="number"&&n.size>0)return`${n.size}px`;if(typeof n.size=="string"){const g=String(n.size).replace(/px/gi,""),c=Number(g);if(c&&!isNaN(c)&&Number.isInteger(c)&&c>0)return`${c}px`}return j}),x=A(()=>({boxSizing:"border-box",color:n.color,display:n.display,flexShrink:"0",height:e.value,lineHeight:"0",width:e.value,pointerEvents:n.decorative?"none":void 0})),N=g=>{const c={},C=Math.random().toString(36).substring(2,12);return g.replace(/id="([^"]+)"/g,(z,$)=>{const E=`${C}-${$}`;return c[$]=E,`id="${E}"`}).replace(/#([^\s^")]+)/g,(z,$)=>c[$]?`#${c[$]}`:z)},t=g=>{const c={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;","`":"&#039;"};return g.replace(/[<>"'`]/g,C=>c[C])},k=`${n.title?`<title data-testid="kui-icon-svg-title">${t(n.title)}</title>`:""}${n.staticIds?nt:N(nt)}`;return(g,c)=>(s(),b(et(d.as),{"aria-hidden":d.decorative?"true":void 0,class:"kui-icon gateway-icon","data-testid":"kui-icon-wrapper-gateway-icon",style:ot(x.value),tabindex:d.decorative?"-1":void 0},{default:i(()=>[(s(),p("svg",{"aria-hidden":d.decorative?"true":void 0,"data-testid":"kui-icon-svg-gateway-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:k},null,8,xt))]),_:1},8,["aria-hidden","style","tabindex"]))}}),wt=["B","kB","MB","GB","TB","PB","EB","ZB","YB"],St=["B","KiB","MiB","GiB","TiB","PiB","EiB","ZiB","YiB"],Ct=["b","kbit","Mbit","Gbit","Tbit","Pbit","Ebit","Zbit","Ybit"],It=["b","kibit","Mibit","Gibit","Tibit","Pibit","Eibit","Zibit","Yibit"],it=(d,n,e)=>{let x=d;return typeof n=="string"||Array.isArray(n)?x=d.toLocaleString(n,e):(n===!0||e!==void 0)&&(x=d.toLocaleString(void 0,e)),x};function Y(d,n){if(!Number.isFinite(d))throw new TypeError(`Expected a finite number, got ${typeof d}: ${d}`);n={bits:!1,binary:!1,space:!0,...n};const e=n.bits?n.binary?It:Ct:n.binary?St:wt,x=n.space?" ":"";if(n.signed&&d===0)return` 0${x}${e[0]}`;const N=d<0,t=N?"-":n.signed?"+":"";N&&(d=-d);let k;if(n.minimumFractionDigits!==void 0&&(k={minimumFractionDigits:n.minimumFractionDigits}),n.maximumFractionDigits!==void 0&&(k={maximumFractionDigits:n.maximumFractionDigits,...k}),d<1){const z=it(d,n.locale,k);return t+z+x+e[0]}const g=Math.min(Math.floor(n.binary?Math.log(d)/Math.log(1024):Math.log10(d)/3),e.length-1);d/=(n.binary?1024:1e3)**g,k||(d=d.toPrecision(3));const c=it(Number(d),n.locale,k),C=e[g];return t+c+x+C}const Bt={},Nt={class:"card"},Tt={class:"title"},qt={class:"body"};function zt(d,n){const e=B("KCard");return s(),b(e,{class:"data-card"},{default:i(()=>[o("dl",null,[o("div",Nt,[o("dt",Tt,[H(d.$slots,"title",{},void 0,!0)]),n[0]||(n[0]=a()),o("dd",qt,[H(d.$slots,"default",{},void 0,!0)])])])]),_:3})}const dt=X(Bt,[["render",zt],["__scopeId","data-v-3f9a3cf3"]]),Dt={class:"title"},Lt={key:0},Vt={"data-testid":"grpc-success"},Et={"data-testid":"grpc-failure"},Mt={"data-testid":"rq-2xx"},ht={"data-testid":"rq-4xx"},Ht={"data-testid":"rq-5xx"},Rt={"data-testid":"connections-total"},Kt={key:0,"data-testid":"bytes-received"},Pt={key:1,"data-testid":"bytes-sent"},At=K({__name:"ConnectionCard",props:{protocol:{},service:{default:""},traffic:{default:void 0},direction:{default:"downstream"}},setup(d){const{t:n}=lt(),e=d,x=N=>{const t=N.target;if(N.isTrusted&&t.nodeName.toLowerCase()!=="a"){const k=t.closest(".service-traffic-card, a");if(k){const g=k.nodeName.toLowerCase()==="a"?k:k.querySelector("[data-action]");g!==null&&"click"in g&&typeof g.click=="function"&&g.click()}}};return(N,t)=>{const k=B("XBadge"),g=B("KSkeletonBox");return s(),b(dt,{class:"service-traffic-card",onClick:x},{title:i(()=>[e.service.length>0?(s(),b(rt,{key:0,tags:[{label:"kuma.io/service",value:e.service}]},null,8,["tags"])):V("",!0),t[1]||(t[1]=a()),o("div",Dt,[u(k,{class:"protocol",appearance:e.protocol==="passthrough"?"success":"info"},{default:i(()=>[a(r(m(n)(`data-planes.components.service_traffic_card.protocol.${e.protocol}`,{},{defaultMessage:m(n)(`http.api.value.${e.protocol}`)})),1)]),_:1},8,["appearance"]),t[0]||(t[0]=a()),H(N.$slots,"default",{},void 0,!0)])]),default:i(()=>{var c,C,z,$,E,U,F,G,O,Z,w,f;return[t[22]||(t[22]=a()),e.traffic?(s(),p("dl",Lt,[e.protocol==="passthrough"?(s(!0),p(_,{key:0},q([["http","tcp"].reduce((S,M)=>{var h;const I=e.direction;return Object.entries(((h=e.traffic)==null?void 0:h[M])||{}).reduce((P,[l,y])=>[`${I}_cx_tx_bytes_total`,`${I}_cx_rx_bytes_total`].includes(l)?{...P,[l]:y+(P[l]??0)}:P,S)},{})],(S,M)=>(s(),p(_,{key:M},[o("div",null,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.tx")),1),t[2]||(t[2]=a()),o("dd",null,r(m(Y)(S.downstream_cx_rx_bytes_total??0)),1)]),t[4]||(t[4]=a()),o("div",null,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.rx")),1),t[3]||(t[3]=a()),o("dd",null,r(m(Y)(S.downstream_cx_tx_bytes_total??0)),1)])],64))),128)):e.protocol==="grpc"?(s(),p(_,{key:1},[o("div",Vt,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.grpc_success")),1),t[5]||(t[5]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:(c=e.traffic.grpc)==null?void 0:c.success})),1)]),t[7]||(t[7]=a()),o("div",Et,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.grpc_failure")),1),t[6]||(t[6]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:(C=e.traffic.grpc)==null?void 0:C.failure})),1)])],64)):e.protocol.startsWith("http")?(s(),p(_,{key:2},[(s(!0),p(_,null,q([((z=e.traffic.http)==null?void 0:z[`${e.direction}_rq_1xx`])??0].filter(S=>S!==0),S=>(s(),p("div",{key:S,"data-testid":"rq-1xx"},[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.1xx")),1),t[8]||(t[8]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:S})),1)]))),128)),t[13]||(t[13]=a()),o("div",Mt,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.2xx")),1),t[9]||(t[9]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:($=e.traffic.http)==null?void 0:$[`${e.direction}_rq_2xx`]})),1)]),t[14]||(t[14]=a()),(s(!0),p(_,null,q([((E=e.traffic.http)==null?void 0:E[`${e.direction}_rq_3xx`])??0].filter(S=>S!==0),S=>(s(),p("div",{key:S,"data-testid":"rq-3xx"},[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.3xx")),1),t[10]||(t[10]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:S})),1)]))),128)),t[15]||(t[15]=a()),o("div",ht,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.4xx")),1),t[11]||(t[11]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:(U=e.traffic.http)==null?void 0:U[`${e.direction}_rq_4xx`]})),1)]),t[16]||(t[16]=a()),o("div",Ht,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.5xx")),1),t[12]||(t[12]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:(F=e.traffic.http)==null?void 0:F[`${e.direction}_rq_5xx`]})),1)])],64)):(s(),p(_,{key:3},[o("div",Rt,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.cx")),1),t[17]||(t[17]=a()),o("dd",null,r(m(n)("common.formats.integer",{value:(G=e.traffic.tcp)==null?void 0:G[`${e.direction}_cx_total`]})),1)]),t[20]||(t[20]=a()),typeof((O=e.traffic.tcp)==null?void 0:O[`${e.direction}_cx_tx_bytes_total`])<"u"?(s(),p("div",Kt,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.rx")),1),t[18]||(t[18]=a()),o("dd",null,r(m(Y)((Z=e.traffic.tcp)==null?void 0:Z[`${e.direction}_cx_tx_bytes_total`])),1)])):V("",!0),t[21]||(t[21]=a()),typeof((w=e.traffic.tcp)==null?void 0:w[`${e.direction}_cx_rx_bytes_total`])<"u"?(s(),p("div",Pt,[o("dt",null,r(m(n)("data-planes.components.service_traffic_card.tx")),1),t[19]||(t[19]=a()),o("dd",null,r(m(Y)((f=e.traffic.tcp)==null?void 0:f[`${e.direction}_cx_rx_bytes_total`])),1)])):V("",!0)],64))])):(s(),b(g,{key:1,width:"10"}))]}),_:3})}}}),W=X(At,[["__scopeId","data-v-82875ef3"]]),Xt={class:"body"},Ut=K({__name:"ConnectionGroup",props:{type:{}},setup(d){const n=d;return(e,x)=>{const N=B("KCard");return s(),b(N,{class:ut(["service-traffic-group",`type-${n.type}`])},{default:i(()=>[o("div",Xt,[H(e.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}}),tt=X(Ut,[["__scopeId","data-v-a119a970"]]),Ft={class:"service-traffic"},Gt={class:"actions"},Ot=K({__name:"ConnectionTraffic",setup(d){return(n,e)=>(s(),p("div",Ft,[o("div",Gt,[H(n.$slots,"actions",{},void 0,!0)]),e[0]||(e[0]=a()),u(dt,{class:"header"},{title:i(()=>[H(n.$slots,"title",{},void 0,!0)]),_:3}),e[1]||(e[1]=a()),H(n.$slots,"default",{},void 0,!0)]))}}),st=X(Ot,[["__scopeId","data-v-e6bd176c"]]),Zt={"data-testid":"dataplane-warnings"},Yt=["data-testid","innerHTML"],jt={key:0,"data-testid":"warning-stats-loading"},Jt={class:"stack","data-testid":"dataplane-details"},Qt={class:"stack"},Wt={class:"columns"},te={class:"status-with-reason"},ee={class:"columns"},ae={class:"columns"},ne={"data-testid":"dataplane-mtls"},ie={class:"columns"},se=["innerHTML"],oe={key:0,"data-testid":"dataplane-subscriptions"},re=K({__name:"DataPlaneDetailView",props:{data:{},mesh:{}},setup(d){const n=pt(),e=d,x=A(()=>e.data.warnings.concat(...e.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(N,t)=>{const k=B("KTooltip"),g=B("DataCollection"),c=B("XAction"),C=B("KCard"),z=B("XEmptyState"),$=B("XLayout"),E=B("XInputSwitch"),U=B("RouterView"),F=B("XAlert"),G=B("AppView"),O=B("DataSource"),Z=B("RouteView");return s(),b(Z,{params:{mesh:"",dataPlane:"",subscription:"",inactive:!1},name:"data-plane-detail-view"},{default:i(({route:w,t:f,can:S,me:M})=>[u(O,{src:`/meshes/${w.params.mesh}/dataplanes/${w.params.dataPlane}/stats/${e.data.dataplane.networking.inboundAddress}`},{default:i(({data:I,error:h,refresh:P})=>[u(G,null,J({default:i(()=>[t[53]||(t[53]=a()),o("div",Jt,[u(C,null,{default:i(()=>[o("div",Qt,[o("div",Wt,[u(L,null,{title:i(()=>[a(r(f("http.api.property.status")),1)]),body:i(()=>[o("div",te,[u(ct,{status:e.data.status},null,8,["status"]),t[4]||(t[4]=a()),e.data.dataplaneType==="standard"?(s(),b(g,{key:0,items:e.data.dataplane.networking.inbounds,predicate:l=>l.state!=="Ready",empty:!1},{default:i(({items:l})=>[u(k,{class:"reason-tooltip"},{content:i(()=>[o("ul",null,[(s(!0),p(_,null,q(l,y=>(s(),p("li",{key:`${y.service}:${y.port}`},r(f("data-planes.routes.item.unhealthy_inbound",{service:y.service,port:y.port})),1))),128))])]),default:i(()=>[u(m(vt),{color:m(mt),size:m(Q)},null,8,["color","size"]),t[3]||(t[3]=a())]),_:2},1024)]),_:2},1032,["items","predicate"])):V("",!0)])]),_:2},1024),t[12]||(t[12]=a()),S("use zones")&&e.data.zone?(s(),b(L,{key:0},{title:i(()=>t[6]||(t[6]=[a(`
                    Zone
                  `)])),body:i(()=>[u(c,{to:{name:"zone-cp-detail-view",params:{zone:e.data.zone}}},{default:i(()=>[a(r(e.data.zone),1)]),_:1},8,["to"])]),_:1})):V("",!0),t[13]||(t[13]=a()),u(L,null,{title:i(()=>t[8]||(t[8]=[a(`
                    Type
                  `)])),body:i(()=>[a(r(f(`data-planes.type.${e.data.dataplaneType}`)),1)]),_:2},1024),t[14]||(t[14]=a()),e.data.namespace.length>0?(s(),b(L,{key:1},{title:i(()=>t[10]||(t[10]=[a(`
                    Namespace
                  `)])),body:i(()=>[a(r(e.data.namespace),1)]),_:1})):V("",!0)]),t[20]||(t[20]=a()),o("div",ee,[u(L,null,{title:i(()=>[a(r(f("data-planes.routes.item.last_updated")),1)]),body:i(()=>[a(r(f("common.formats.datetime",{value:Date.parse(e.data.modificationTime)})),1)]),_:2},1024),t[19]||(t[19]=a()),e.data.dataplane.networking.gateway?(s(),p(_,{key:0},[u(L,null,{title:i(()=>[a(r(f("http.api.property.tags")),1)]),body:i(()=>[u(rt,{tags:e.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),t[18]||(t[18]=a()),u(L,null,{title:i(()=>[a(r(f("http.api.property.address")),1)]),body:i(()=>[u(ft,{text:`${e.data.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)],64)):V("",!0)])])]),_:2},1024),t[49]||(t[49]=a()),u(C,{class:"traffic","data-testid":"dataplane-traffic"},{default:i(()=>[o("div",ae,[u(st,null,{title:i(()=>[u(m(kt),{display:"inline-block",decorative:"",size:m(Q)},null,8,["size"]),t[21]||(t[21]=a(`
                  Inbounds
                `))]),default:i(()=>[t[23]||(t[23]=a()),u(tt,{type:"inbound","data-testid":"dataplane-inbounds"},{default:i(()=>[(s(!0),p(_,null,q([e.data.dataplane.networking.type==="gateway"?Object.entries((I==null?void 0:I.inbounds)??{}).reduce((l,[y,v])=>{var D;const T=y.split("_").at(-1);return T===(((D=e.data.dataplane.networking.admin)==null?void 0:D.port)??"9901")?l:l.concat([{...e.data.dataplane.networking.inbounds[0],name:y,port:Number(T),protocol:["http","tcp"].find(R=>typeof v[R]<"u")??"tcp",addressPort:`${e.data.dataplane.networking.inbounds[0].address}:${T}`}])},[]):e.data.dataplane.networking.inbounds],l=>(s(),b(g,{key:l,items:l,predicate:y=>y.port!==49151},J({default:i(({items:y})=>[u($,{type:"stack",size:"small"},{default:i(()=>[(s(!0),p(_,null,q(y,v=>(s(),p(_,{key:`${v.name}`},[(s(!0),p(_,null,q([I==null?void 0:I.inbounds[v.name]],T=>(s(),b(W,{key:T,"data-testid":"dataplane-inbound",protocol:v.protocol,service:S("use service-insights",e.mesh)?v.tags["kuma.io/service"]:"",traffic:typeof h>"u"?T:{name:"",protocol:v.protocol,port:`${v.port}`}},{default:i(()=>[u(c,{"data-action":"",to:{name:(D=>D.includes("bound")?D.replace("-outbound-","-inbound-"):"connection-inbound-summary-overview-view")(String(m(n).name)),params:{connection:v.name},query:{inactive:w.params.inactive}}},{default:i(()=>[a(r(v.name.replace("localhost","").replace("_",":")),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","service","traffic"]))),128))],64))),128))]),_:2},1024)]),_:2},[e.data.dataplaneType==="delegated"?{name:"empty",fn:i(()=>[u(z,null,{default:i(()=>[o("p",null,`
                            This proxy is a delegated gateway therefore `+r(f("common.product.name"))+` does not have any visibility into inbounds for this gateway.
                          `,1)]),_:2},1024)]),key:"0"}:void 0]),1032,["items","predicate"]))),128))]),_:2},1024)]),_:2},1024),t[33]||(t[33]=a()),u(st,null,J({title:i(()=>[u(m($t),{display:"inline-block",decorative:"",size:m(Q)},null,8,["size"]),t[27]||(t[27]=a()),t[28]||(t[28]=o("span",null,"Outbounds",-1))]),default:i(()=>[t[31]||(t[31]=a()),t[32]||(t[32]=a()),typeof h>"u"?(s(),p(_,{key:0},[typeof I>"u"?(s(),b(yt,{key:0})):(s(),p(_,{key:1},q(["upstream"],l=>(s(),p(_,{key:l},[u(tt,{type:"passthrough"},{default:i(()=>[u(W,{protocol:"passthrough",traffic:I.passthrough},{default:i(()=>t[29]||(t[29]=[a(`
                          Non mesh traffic
                        `)])),_:2},1032,["traffic"])]),_:2},1024),t[30]||(t[30]=a()),u(g,{predicate:w.params.inactive?void 0:([y,v])=>{var T,D;return((typeof v.tcp<"u"?(T=v.tcp)==null?void 0:T[`${l}_cx_rx_bytes_total`]:(D=v.http)==null?void 0:D[`${l}_rq_total`])??0)>0},items:Object.entries(I.outbounds)},{default:i(({items:y})=>[y.length>0?(s(),b(tt,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:i(()=>[(s(),p(_,null,q([/-([a-f0-9]){16}$/],v=>u($,{key:v,type:"stack",size:"small"},{default:i(()=>[(s(!0),p(_,null,q(y,([T,D])=>(s(),b(W,{key:`${T}`,"data-testid":"dataplane-outbound",protocol:["grpc","http","tcp"].find(R=>typeof D[R]<"u")??"tcp",traffic:D,service:D.$resourceMeta.type===""?T.replace(v,""):void 0,direction:l},{default:i(()=>[u(c,{"data-action":"",to:{name:(R=>R.includes("bound")?R.replace("-inbound-","-outbound-"):"connection-outbound-summary-overview-view")(String(m(n).name)),params:{connection:T},query:{inactive:w.params.inactive}}},{default:i(()=>[a(r(T),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic","service","direction"]))),128))]),_:2},1024)),64))]),_:2},1024)):V("",!0)]),_:2},1032,["predicate","items"])],64))),64))],64)):(s(),b(z,{key:1}))]),_:2},[I?{name:"actions",fn:i(()=>[u(E,{modelValue:w.params.inactive,"onUpdate:modelValue":l=>w.params.inactive=l,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:i(()=>t[24]||(t[24]=[a(`
                      Show inactive
                    `)])),_:2},1032,["modelValue","onUpdate:modelValue"]),t[26]||(t[26]=a()),u(c,{action:"refresh",appearance:"primary",onClick:P},{default:i(()=>t[25]||(t[25]=[a(`
                    Refresh
                  `)])),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)])]),_:2},1024),t[50]||(t[50]=a()),u(U,null,{default:i(l=>[l.route.name!==w.name?(s(),b(_t,{key:0,width:"670px",onClose:function(){w.replace({name:"data-plane-detail-view",params:{mesh:w.params.mesh,dataPlane:w.params.dataPlane},query:{inactive:w.params.inactive?null:void 0}})}},{default:i(()=>[(s(),b(et(l.Component),{data:w.params.subscription.length>0?e.data.dataplaneInsight.subscriptions:l.route.name.includes("-inbound-")?e.data.dataplane.networking.inbounds:(I==null?void 0:I.outbounds)||{},"dataplane-overview":e.data},null,8,["data","dataplane-overview"]))]),_:2},1032,["onClose"])):V("",!0)]),_:2},1024),t[51]||(t[51]=a()),o("div",ne,[o("h2",null,r(f("data-planes.routes.item.mtls.title")),1),t[43]||(t[43]=a()),e.data.dataplaneInsight.mTLS?(s(!0),p(_,{key:0},q([e.data.dataplaneInsight.mTLS],l=>(s(),b(C,{key:l,class:"mt-4"},{default:i(()=>[o("div",ie,[u(L,null,{title:i(()=>[a(r(f("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:i(()=>[a(r(f("common.formats.datetime",{value:Date.parse(l.certificateExpirationTime)})),1)]),_:2},1024),t[39]||(t[39]=a()),u(L,null,{title:i(()=>[a(r(f("data-planes.routes.item.mtls.generation_time.title")),1)]),body:i(()=>[a(r(f("common.formats.datetime",{value:Date.parse(l.lastCertificateRegeneration)})),1)]),_:2},1024),t[40]||(t[40]=a()),u(L,null,{title:i(()=>[a(r(f("data-planes.routes.item.mtls.regenerations.title")),1)]),body:i(()=>[a(r(f("common.formats.integer",{value:l.certificateRegenerations})),1)]),_:2},1024),t[41]||(t[41]=a()),u(L,null,{title:i(()=>[a(r(f("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:i(()=>[a(r(l.issuedBackend),1)]),_:2},1024),t[42]||(t[42]=a()),u(L,null,{title:i(()=>[a(r(f("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:i(()=>[o("ul",null,[(s(!0),p(_,null,q(l.supportedBackends,y=>(s(),p("li",{key:y},r(y),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(s(),b(F,{key:1,class:"mt-4",appearance:"warning"},{default:i(()=>[o("div",{innerHTML:f("data-planes.routes.item.mtls.disabled")},null,8,se)]),_:2},1024))]),t[52]||(t[52]=a()),e.data.dataplaneInsight.subscriptions.length>0?(s(),p("div",oe,[o("h2",null,r(f("data-planes.routes.item.subscriptions.title")),1),t[48]||(t[48]=a()),u(gt,{headers:[{...M.get("headers.instanceId"),label:f("http.api.property.instanceId"),key:"instanceId"},{...M.get("headers.version"),label:f("http.api.property.version"),key:"version"},{...M.get("headers.connected"),label:f("http.api.property.connected"),key:"connected"},{...M.get("headers.disconnected"),label:f("http.api.property.disconnected"),key:"disconnected"},{...M.get("headers.responses"),label:f("http.api.property.responses"),key:"responses"}],"is-selected-row":l=>l.id===w.params.subscription,items:e.data.dataplaneInsight.subscriptions.map((l,y,v)=>v[v.length-(y+1)]),onResize:M.set},{instanceId:i(({row:l})=>[u(c,{"data-action":"",to:{name:"data-plane-subscription-summary-view",params:{subscription:l.id}}},{default:i(()=>[a(r(l.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:i(({row:l})=>{var y,v;return[a(r(((v=(y=l.version)==null?void 0:y.kumaDp)==null?void 0:v.version)??"-"),1)]}),connected:i(({row:l})=>[a(r(f("common.formats.datetime",{value:Date.parse(l.connectTime??"")})),1)]),disconnected:i(({row:l})=>[l.disconnectTime?(s(),p(_,{key:0},[a(r(f("common.formats.datetime",{value:Date.parse(l.disconnectTime)})),1)],64)):V("",!0)]),responses:i(({row:l})=>{var y;return[(s(!0),p(_,null,q([((y=l.status)==null?void 0:y.total)??{}],v=>(s(),p(_,null,[a(r(v.responsesSent)+"/"+r(v.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"])])):V("",!0)])]),_:2},[x.value.length>0||h?{name:"notifications",fn:i(()=>[o("ul",Zt,[(s(!0),p(_,null,q(x.value,l=>(s(),p("li",{key:l.kind,"data-testid":`warning-${l.kind}`,innerHTML:f(`common.warnings.${l.kind}`,l.payload)},null,8,Yt))),128)),t[2]||(t[2]=a()),h?(s(),p("li",jt,[t[0]||(t[0]=a(`
              The below view is not enhanced with runtime stats (Error loading stats: `)),o("strong",null,r(h.toString()),1),t[1]||(t[1]=a(`)
            `))])):V("",!0)])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}}),ce=X(re,[["__scopeId","data-v-a77ac2ac"]]);export{ce as default};
