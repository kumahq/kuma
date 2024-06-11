import{d as L,I as F,c as y,o as l,a as g,w as r,b as _,t as s,e as w,n as pe,r as _e,p as fe,f as ve,g as o,s as be,z as A,F as j,k as t,A as n,j as a,R as ze,G as z,S as D,i as p,H as N,J as ge,T as H,U as W,V as G,W as he,_ as ke}from"./index-DQUwSwHF.js";import{m as ye}from"./kong-icons.es245-CCTPTDJC.js";import{m as Ce}from"./kong-icons.es265-YFtev5QL.js";import{C}from"./CodeBlock-DCsN1aa-.js";const Ee=c=>(fe("data-v-81eeb177"),c=c(),ve(),c),we=["aria-hidden"],Se={key:0,"data-testid":"kui-icon-svg-title"},$e=Ee(()=>o("path",{d:"M10.6 16.6L17.65 9.55L16.25 8.15L10.6 13.8L7.75 10.95L6.35 12.35L10.6 16.6ZM12 22C10.6167 22 9.31667 21.7375 8.1 21.2125C6.88333 20.6875 5.825 19.975 4.925 19.075C4.025 18.175 3.3125 17.1167 2.7875 15.9C2.2625 14.6833 2 13.3833 2 12C2 10.6167 2.2625 9.31667 2.7875 8.1C3.3125 6.88333 4.025 5.825 4.925 4.925C5.825 4.025 6.88333 3.3125 8.1 2.7875C9.31667 2.2625 10.6167 2 12 2C13.3833 2 14.6833 2.2625 15.9 2.7875C17.1167 3.3125 18.175 4.025 19.075 4.925C19.975 5.825 20.6875 6.88333 21.2125 8.1C21.7375 9.31667 22 10.6167 22 12C22 13.3833 21.7375 14.6833 21.2125 15.9C20.6875 17.1167 19.975 18.175 19.075 19.075C18.175 19.975 17.1167 20.6875 15.9 21.2125C14.6833 21.7375 13.3833 22 12 22Z",fill:"currentColor"},null,-1)),Ne=L({__name:"CheckCircleIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:F,validator:c=>{if(typeof c=="number"&&c>0)return!0;if(typeof c=="string"){const e=String(c).replace(/px/gi,""),m=Number(e);if(m&&!isNaN(m)&&Number.isInteger(m)&&m>0)return!0}return!1}},as:{type:String,required:!1,default:"span"}},setup(c){const e=c,m=y(()=>{if(typeof e.size=="number"&&e.size>0)return`${e.size}px`;if(typeof e.size=="string"){const E=String(e.size).replace(/px/gi,""),u=Number(E);if(u&&!isNaN(u)&&Number.isInteger(u)&&u>0)return`${u}px`}return F}),f=y(()=>({boxSizing:"border-box",color:e.color,display:e.display,height:m.value,lineHeight:"0",width:m.value}));return(E,u)=>(l(),g(_e(c.as),{"aria-hidden":c.decorative?"true":void 0,class:"kui-icon check-circle-icon","data-testid":"kui-icon-wrapper-check-circle-icon",style:pe(f.value)},{default:r(()=>[(l(),_("svg",{"aria-hidden":c.decorative?"true":void 0,"data-testid":"kui-icon-svg-check-circle-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg"},[c.title?(l(),_("title",Se,s(c.title),1)):w("",!0),$e],8,we))]),_:1},8,["aria-hidden","style"]))}}),Ie=be(Ne,[["__scopeId","data-v-81eeb177"]]),Ke={class:"form-step-title"},Ve=o("span",{class:"form-step-number"},"1",-1),Le={class:"instruction-list"},Te={href:"https://helm.sh/docs/intro/install/"},qe={class:"form-step-title"},Ze=o("span",{class:"form-step-number"},"2",-1),xe=o("p",null,"On your local machine, create a namespace in your Kubernetes cluster and pull down the kong Helm repo.",-1),De={class:"instruction-list"},Ae={class:"form-step-title"},Be=o("span",{class:"form-step-number"},"3",-1),Re={class:"form-step-title"},Ue=o("span",{class:"form-step-number"},"4",-1),Pe={class:"field-group-label mt-4"},Xe={class:"mt-4"},Me=L({__name:"ZoneCreateKubernetesInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},zoneIngressEnabled:{type:Boolean,required:!0},zoneEgressEnabled:{type:Boolean,required:!0},token:{type:String,required:!0},base64EncodedToken:{type:String,required:!0}},setup(c){const{t:e}=A(),m=j(),f=c,E=y(()=>e("zones.form.kubernetes.secret.createSecretCommand",{token:f.base64EncodedToken}).trim()),u=y(()=>{const h={zoneName:f.zoneName,globalKdsAddress:f.globalKdsAddress,zoneIngressEnabled:String(f.zoneIngressEnabled),zoneEgressEnabled:String(f.zoneEgressEnabled),controlPlaneId:typeof m.params.virtualControlPlaneId=="string"?m.params.virtualControlPlaneId:""};return e("zones.form.kubernetes.connectZone.config",h).trim()});return(h,v)=>(l(),_("div",null,[o("h3",Ke,[Ve,t(" "+s(n(e)("zones.form.kubernetes.prerequisites.title")),1)]),t(),o("ul",Le,[o("li",null,[o("b",null,s(n(e)("zones.form.kubernetes.prerequisites.step1Label"))+s(f.zoneIngressEnabled?" "+n(e)("zones.form.kubernetes.prerequisites.step1LabelAddendum"):""),1),t(`:
        `+s(n(e)("zones.form.kubernetes.prerequisites.step1Description",{productName:n(e)("common.product.name")})),1)]),t(),o("li",null,[o("b",null,s(n(e)("zones.form.kubernetes.prerequisites.step2Label")),1),t(`:
        `+s(n(e)("zones.form.kubernetes.prerequisites.step2Description")),1)]),t(),o("li",null,[o("a",Te,s(n(e)("zones.form.kubernetes.prerequisites.step3LinkTitle")),1),t(" "+s(n(e)("zones.form.kubernetes.prerequisites.step3Tail")),1)])]),t(),o("h3",qe,[Ze,t(" "+s(n(e)("zones.form.kubernetes.helm.title")),1)]),t(),xe,t(),o("ol",De,[o("li",null,[o("b",null,s(n(e)("zones.form.kubernetes.helm.step1Description")),1),t(),a(C,{class:"mt-2",code:n(e)("zones.form.kubernetes.helm.step1Command"),language:"bash"},null,8,["code"])]),t(),o("li",null,[o("b",null,s(n(e)("zones.form.kubernetes.helm.step2Description")),1),t(),a(C,{class:"mt-2",code:n(e)("zones.form.kubernetes.helm.step2Command"),language:"bash"},null,8,["code"])]),t(),o("li",null,[o("b",null,s(n(e)("zones.form.kubernetes.helm.step3Description")),1),t(),a(C,{class:"mt-2",code:n(e)("zones.form.kubernetes.helm.step3Command"),language:"bash"},null,8,["code"])])]),t(),o("h3",Ae,[Be,t(" "+s(n(e)("zones.form.kubernetes.secret.title")),1)]),t(),o("p",null,s(n(e)("zones.form.kubernetes.secret.createSecretDescription")),1),t(),a(C,{class:"mt-4",code:E.value,language:"bash"},null,8,["code"]),t(),o("h3",Re,[Ue,t(" "+s(n(e)("zones.form.kubernetes.connectZone.title")),1)]),t(),o("p",null,s(n(e)("zones.form.kubernetes.connectZone.configDescription")),1),t(),o("span",Pe,s(n(e)("zones.form.kubernetes.connectZone.configFileName")),1),t(),a(C,{"data-testid":"zone-kubernetes-config",code:u.value,language:"yaml"},null,8,["code"]),t(),o("p",Xe,s(n(e)("zones.form.kubernetes.connectZone.connectDescription")),1),t(),a(C,{class:"mt-4",code:n(e)("zones.form.kubernetes.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),Oe={class:"form-step-title"},Fe=o("span",{class:"form-step-number"},"1",-1),He={class:"form-step-title"},We=o("span",{class:"form-step-number"},"2",-1),Ge={class:"field-group-label mt-4"},je={class:"mt-4"},Je=L({__name:"ZoneCreateUniversalInstructions",props:{zoneName:{type:String,required:!0},globalKdsAddress:{type:String,required:!0},token:{type:String,required:!0}},setup(c){const{t:e}=A(),m=j(),f=c,E=y(()=>e("zones.form.universal.saveToken.saveTokenCommand",{token:f.token}).trim()),u=y(()=>{const h={zoneName:f.zoneName,globalKdsAddress:f.globalKdsAddress,controlPlaneId:typeof m.params.virtualControlPlaneId=="string"?m.params.virtualControlPlaneId:""};return e("zones.form.universal.connectZone.config",h).trim()});return(h,v)=>(l(),_("div",null,[o("h3",Oe,[Fe,t(" "+s(n(e)("zones.form.universal.saveToken.title")),1)]),t(),o("p",null,s(n(e)("zones.form.universal.saveToken.saveTokenDescription")),1),t(),a(C,{class:"mt-4",code:E.value,language:"bash"},null,8,["code"]),t(),o("h3",He,[We,t(" "+s(n(e)("zones.form.universal.connectZone.title")),1)]),t(),o("p",null,s(n(e)("zones.form.universal.connectZone.configDescription")),1),t(),o("span",Ge,s(n(e)("zones.form.universal.connectZone.configFileName")),1),t(),a(C,{"data-testid":"zone-universal-config",class:"mt-4",code:u.value,language:"yaml"},null,8,["code"]),t(),o("p",je,s(n(e)("zones.form.universal.connectZone.connectDescription")),1),t(),a(C,{class:"mt-4",code:n(e)("zones.form.universal.connectZone.connectCommand").trim(),language:"bash"},null,8,["code"])]))}}),Qe={class:"form-wrapper"},Ye={key:1},et={key:2},tt={class:"form"},ot={class:"form-header"},nt={class:"form-title"},st={class:"text-gradient"},at={key:0},rt={key:0},lt={class:"fact-list"},it={class:"form-section"},ct={class:"form-section__header"},ut={class:"form-section-title"},dt={class:"form-section__content"},mt={class:"form-section","data-testid":"connect-zone-instructions"},pt={class:"form-section__header"},_t={class:"form-section-title"},ft={class:"form-section__content"},vt={class:"field-group-list"},bt={class:"field-group"},zt={class:"field-group-label"},gt={class:"radio-button-group"},ht={class:"field-group"},kt={class:"field-group-label"},yt={class:"radio-button-group","data-testid":"ingress-input-switch"},Ct={class:"field-group"},Et={class:"field-group-label"},wt={class:"radio-button-group","data-testid":"egress-input-switch"},St={class:"form-section"},$t={class:"form-section__header"},Nt={class:"form-section-title"},It={class:"form-section__content"},Kt={class:"form-section"},Vt={class:"form-section__header"},Lt={class:"form-section-title"},Tt={class:"form-section__content"},qt={class:"mt-2"},Zt=L({__name:"ZoneCreateView",setup(c){const{t:e,tm:m}=A(),f=ze(),E=/^(?![-0-9])[a-z0-9-]{1,63}$/,u=z(null),h=z(!1),v=z(null),K=z(null),T=z(""),B=z(!1),J=z(new Date),b=z(""),$=z("kubernetes"),q=z(!0),Z=z(!0),I=y(()=>u.value!==null&&u.value.token?u.value.token:""),Q=y(()=>I.value!==""?window.btoa(I.value):""),Y=y(()=>b.value===""||h.value||u.value!==null),x=y(()=>{if(K.value!==null)return K.value;if(v.value instanceof D){const k=v.value.invalidParameters.find(d=>d.field==="name");if(k!==void 0)return k.reason}return null});async function ee(){h.value=!0,v.value=null,T.value="";try{if(!R(b.value))return;u.value=await f.createZone({name:b.value})}catch(k){k instanceof Error?(T.value=b.value,v.value=k):console.error(k)}finally{h.value=!1}}function R(k){const d=E.test(k);return d?K.value=null:K.value=e("zones.create.invalidNameError"),d}function te(){B.value=!0}return(k,d)=>{const oe=p("RouteTitle"),V=p("KButton"),ne=p("KModal"),se=p("XTeleportTemplate"),ae=p("XDisclosure"),re=p("KAlert"),le=p("KLabel"),ie=p("KInput"),U=p("KRadio"),P=p("KInputSwitch"),X=p("DataSource"),ce=p("KEmptyState"),ue=p("KCard"),de=p("AppView"),me=p("RouteView");return l(),g(me,{name:"zone-create-view",attrs:{class:"is-fullscreen"}},{default:r(({route:M,id:O})=>[a(de,{fullscreen:!0,breadcrumbs:[]},{title:r(()=>[o("h1",null,[a(oe,{title:n(e)("zones.routes.create.title")},null,8,["title"])])]),actions:r(()=>[a(ae,null,{default:r(({expanded:i,toggle:S})=>[a(V,{appearance:"tertiary","data-testid":"exit-button",onClick:()=>{I.value===""||B.value?M.back({name:"zone-cp-list-view"}):S()}},{default:r(()=>[t(s(n(e)("zones.form.exit")),1)]),_:2},1032,["onClick"]),t(),a(se,{to:{name:"modal-layer"}},{default:r(()=>[a(ne,{visible:i,title:n(e)("zones.form.confirm_modal.title"),"data-testid":"confirm-exit-modal",onCancel:S,onProceed:xt=>M.replace({name:"zone-cp-list-view"})},{"footer-actions":r(()=>[a(V,{appearance:"primary",to:{name:"zone-cp-list-view"},"data-testid":"confirm-exit-button"},{default:r(()=>[t(s(n(e)("zones.form.confirm_modal.action_button")),1)]),_:1})]),default:r(()=>[t(s(n(e)("zones.form.confirm_modal.body"))+" ",1)]),_:2},1032,["visible","title","onCancel","onProceed"])]),_:2},1024)]),_:2},1024)]),default:r(()=>[t(),t(),o("div",Qe,[v.value!==null?(l(),g(re,{key:0,appearance:"danger",class:"mb-4","data-testid":"create-zone-error"},{default:r(()=>[v.value instanceof n(D)&&[409,500].includes(v.value.status)?(l(),_(N,{key:0},[o("p",null,s(n(e)(`zones.create.status_error.${v.value.status}.title`,{name:T.value})),1),t(),o("p",null,s(n(e)(`zones.create.status_error.${v.value.status}.description`)),1)],64)):v.value instanceof n(D)?(l(),_("p",Ye,s(n(e)("common.error_state.api_error",{status:v.value.status,title:v.value.detail})),1)):(l(),_("p",et,s(n(e)("common.error_state.default_error")),1))]),_:1})):w("",!0),t(),a(ue,{class:"form-card"},{default:r(()=>[o("div",tt,[o("div",ot,[o("div",null,[o("h1",nt,[o("span",st,s(n(e)("zones.form.title")),1)]),t(),n(e)("zones.form.description")!==" "?(l(),_("p",at,s(n(e)("zones.form.description")),1)):w("",!0)]),t(),n(m)("zones.form.facts").length>0?(l(),_("div",rt,[o("ul",lt,[(l(!0),_(N,null,ge(n(m)("zones.form.facts"),(i,S)=>(l(),_("li",{key:S,class:"fact-list__item"},[a(n(Ce),{class:"fact-list__icon",color:n(H)},null,8,["color"]),t(" "+s(i),1)]))),128))])])):w("",!0)]),t(),o("div",it,[o("div",ct,[o("h2",ut,s(n(e)("zones.form.section.name.title")),1),t(),o("p",null,s(n(e)("zones.form.section.name.description")),1)]),t(),o("div",dt,[o("div",null,[a(le,{for:O,required:"","tooltip-attributes":{placement:"right"}},{tooltip:r(()=>[t(s(n(e)("zones.form.name_tooltip")),1)]),default:r(()=>[t(s(n(e)("zones.form.nameLabel"))+" ",1)]),_:2},1032,["for"]),t(),a(ie,{id:O,modelValue:b.value,"onUpdate:modelValue":d[0]||(d[0]=i=>b.value=i),type:"text",name:"zone-name","data-testid":"name-input","data-test-error-type":x.value!==null?"invalid-dns-name":void 0,error:x.value!==null,"error-message":x.value??void 0,disabled:u.value!==null,onBlur:d[1]||(d[1]=i=>R(b.value))},null,8,["id","modelValue","data-test-error-type","error","error-message","disabled"])]),t(),a(V,{appearance:"primary",class:"mt-4",disabled:Y.value,"data-testid":"create-zone-button",onClick:ee},{default:r(()=>[h.value?(l(),g(n(W),{key:0,color:n(G)},null,8,["color"])):(l(),g(n(ye),{key:1})),t(" "+s(n(e)("zones.form.createZoneButtonLabel")),1)]),_:1},8,["disabled"])])]),t(),u.value!==null?(l(),_(N,{key:0},[o("div",mt,[o("div",pt,[o("h2",_t,s(n(e)("zones.form.section.configuration.title")),1),t(),o("p",null,s(n(e)("zones.form.section.configuration.description")),1)]),t(),o("div",ft,[o("div",vt,[o("div",bt,[o("span",zt,s(n(e)("zones.form.environmentLabel"))+` *
                      `,1),t(),o("div",gt,[a(U,{modelValue:$.value,"onUpdate:modelValue":d[2]||(d[2]=i=>$.value=i),"selected-value":"universal",name:"zone-environment","data-testid":"environment-universal-radio-button"},{default:r(()=>[t(s(n(e)("zones.form.universalLabel")),1)]),_:1},8,["modelValue"]),t(),a(U,{modelValue:$.value,"onUpdate:modelValue":d[3]||(d[3]=i=>$.value=i),"selected-value":"kubernetes",name:"zone-environment","data-testid":"environment-kubernetes-radio-button"},{default:r(()=>[t(s(n(e)("zones.form.kubernetesLabel")),1)]),_:1},8,["modelValue"])])]),t(),$.value==="kubernetes"?(l(),_(N,{key:0},[o("div",ht,[o("span",kt,s(n(e)("zones.form.zoneIngressLabel"))+` *
                        `,1),t(),o("div",yt,[a(P,{modelValue:q.value,"onUpdate:modelValue":d[4]||(d[4]=i=>q.value=i)},{label:r(()=>[t(s(n(e)("zones.form.zoneIngressEnabledLabel")),1)]),_:1},8,["modelValue"])])]),t(),o("div",Ct,[o("span",Et,s(n(e)("zones.form.zoneEgressLabel"))+` *
                        `,1),t(),o("div",wt,[a(P,{modelValue:Z.value,"onUpdate:modelValue":d[5]||(d[5]=i=>Z.value=i)},{label:r(()=>[t(s(n(e)("zones.form.zoneEgressEnabledLabel")),1)]),_:1},8,["modelValue"])])])],64)):w("",!0)])])]),t(),o("div",St,[o("div",$t,[o("h2",Nt,s(n(e)("zones.form.section.connect_zone.title")),1),t(),o("p",null,s(n(e)("zones.form.section.connect_zone.description")),1)]),t(),o("div",It,[a(X,{src:"/control-plane/addresses"},{default:r(({data:i})=>[typeof i<"u"?(l(),_(N,{key:0},[$.value==="universal"?(l(),g(Je,{key:0,"zone-name":b.value,token:I.value,"global-kds-address":i.kds},null,8,["zone-name","token","global-kds-address"])):(l(),g(Me,{key:1,"zone-name":b.value,"zone-ingress-enabled":q.value,"zone-egress-enabled":Z.value,token:I.value,"base64-encoded-token":Q.value,"global-kds-address":i.kds},null,8,["zone-name","zone-ingress-enabled","zone-egress-enabled","token","base64-encoded-token","global-kds-address"]))],64)):w("",!0)]),_:1})])]),t(),o("div",Kt,[a(X,{src:`/zone-cps/online/${b.value}?no-cache=${J.value}`,onChange:te},{default:r(({data:i,error:S})=>[o("div",Vt,[o("h2",Lt,s(n(e)("zones.form.section.scanner.title")),1),t(),o("p",null,s(n(e)("zones.form.section.scanner.description")),1)]),t(),o("div",Tt,[S?(l(),g(he,{key:0,error:S,appearance:"danger","data-testid":"error"},null,8,["error"])):(l(),g(ce,{key:1},{icon:r(()=>[i===void 0?(l(),g(n(W),{key:0,"data-testid":"waiting",color:n(G)},null,8,["color"])):(l(),g(n(Ie),{key:1,"data-testid":"connected",color:n(H)},null,8,["color"]))]),title:r(()=>[t(s(i===void 0?n(e)("zones.form.scan.waitTitle"):n(e)("zones.form.scan.completeTitle")),1)]),default:r(()=>[t(),t(),typeof i<"u"?(l(),_(N,{key:0},[o("p",null,s(n(e)("zones.form.scan.completeDescription",{name:b.value})),1),t(),o("p",qt,[a(V,{appearance:"primary",to:{name:"zone-cp-detail-view",params:{zone:b.value}}},{default:r(()=>[t(s(n(e)("zones.form.scan.completeButtonLabel",{name:b.value})),1)]),_:1},8,["to"])])],64)):w("",!0)]),_:2},1024))])]),_:1},8,["src"])])],64)):w("",!0)])]),_:2},1024)])]),_:2},1024)]),_:1})}}}),Ut=ke(Zt,[["__scopeId","data-v-56a4ce4f"]]);export{Ut as default};
